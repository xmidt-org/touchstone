// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/httpaux/client"
	"github.com/xmidt-org/touchstone"
)

// newTestFactory creates an isolated touchstone.Factory backed by a fresh registry.
func newTestFactory(t *testing.T) *touchstone.Factory {
	t.Helper()

	_, r, err := touchstone.New(touchstone.Config{})
	require.NoError(t, err)

	return touchstone.NewFactory(touchstone.Config{}, nil, r)
}

// clock is a controllable time source for deterministic duration metrics.
type clock struct {
	now time.Time
}

func (c *clock) Now() time.Time { return c.now }

func (c *clock) Advance(d time.Duration) { c.now = c.now.Add(d) }

// observed extracts the sample count and sum from a single histogram or
// summary within an ObserverVec.
func observed(t *testing.T, ov prometheus.ObserverVec, l prometheus.Labels) (count uint64, sum float64) {
	t.Helper()

	m, ok := ov.With(l).(prometheus.Metric)
	require.True(t, ok, "observer is not a prometheus.Metric")

	var out dto.Metric
	require.NoError(t, m.Write(&out))

	switch {
	case out.Histogram != nil:
		return out.Histogram.GetSampleCount(), out.Histogram.GetSampleSum()
	case out.Summary != nil:
		return out.Summary.GetSampleCount(), out.Summary.GetSampleSum()
	default:
		require.Fail(t, "metric is neither a histogram nor a summary")
		return 0, 0
	}
}

func TestServerInstrumenterThen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		body           string
		handler        http.HandlerFunc
		expectedCode   string
		expectedMethod string
	}{
		{
			name:   "ImplicitOK",
			method: http.MethodGet,
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				_, _ = rw.Write([]byte("ok"))
			},
			expectedCode:   "200",
			expectedMethod: http.MethodGet,
		},
		{
			name:   "ExplicitCode",
			method: http.MethodPost,
			body:   "payload",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			},
			expectedCode:   "404",
			expectedMethod: http.MethodPost,
		},
		{
			name:           "NoWrite",
			method:         http.MethodDelete,
			handler:        func(http.ResponseWriter, *http.Request) {},
			expectedCode:   "200",
			expectedMethod: http.MethodDelete,
		},
		{
			name:   "UnrecognizedMethod",
			method: "BREW",
			handler: func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusTeapot)
			},
			expectedCode:   "418",
			expectedMethod: MethodUnrecognized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				f = newTestFactory(t)
				c = &clock{now: time.Now()}
			)

			si, err := ServerBundle{Now: c.Now}.NewInstrumenter(ServerLabel, "main")(f)
			require.NoError(t, err)

			handler := si.Then(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				assert.InDelta(t, 1.0, testutil.ToFloat64(si.inFlight), 0, "in-flight during handling")
				c.Advance(250 * time.Millisecond)
				tc.handler(rw, r)
			}))

			request := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			handler.ServeHTTP(httptest.NewRecorder(), request)

			l := prometheus.Labels{CodeLabel: tc.expectedCode, MethodLabel: tc.expectedMethod}
			assert.InDelta(t, 0.0, testutil.ToFloat64(si.inFlight), 0, "in-flight after handling")
			assert.InDelta(t, 1.0, testutil.ToFloat64(si.count.With(l)), 0)

			count, sum := observed(t, si.duration, l)
			assert.Equal(t, uint64(1), count)
			assert.InDelta(t, 250.0, sum, 0, "duration in milliseconds")

			count, sum = observed(t, si.requestSize, l)
			assert.Equal(t, uint64(1), count)
			assert.InDelta(t, float64(len(tc.body)), sum, 0, "request size in bytes")

			assert.Nil(t, si.errorCount, "servers do not track errors")
		})
	}
}

func TestClientInstrumenterThen(t *testing.T) {
	t.Parallel()

	var (
		errBoom = errors.New("boom")
		body    = "12345"
	)

	tests := []struct {
		name           string
		response       *http.Response
		err            error
		expectedCode   string
		expectedErrors float64
	}{
		{
			name:         "Success",
			response:     &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody},
			expectedCode: "201",
		},
		{
			name:           "Error",
			err:            errBoom,
			expectedCode:   statusUnknown,
			expectedErrors: 1,
		},
		{
			name:           "ResponseAndError",
			response:       &http.Response{StatusCode: http.StatusBadGateway, Body: http.NoBody},
			err:            errBoom,
			expectedCode:   "502",
			expectedErrors: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				f = newTestFactory(t)
				c = &clock{now: time.Now()}
			)

			ci, err := ClientBundle{Now: c.Now}.NewInstrumenter(ClientLabel, "main")(f)
			require.NoError(t, err)

			decorated := ci.Then(client.Func(func(*http.Request) (*http.Response, error) {
				assert.InDelta(t, 1.0, testutil.ToFloat64(ci.inFlight), 0, "in-flight during transaction")
				c.Advance(100 * time.Millisecond)
				return tc.response, tc.err
			}))

			request := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(body))
			response, err := decorated.Do(request)
			if response != nil {
				response.Body.Close()
			}

			assert.Equal(t, tc.response, response)
			assert.Equal(t, tc.err, err)

			l := prometheus.Labels{CodeLabel: tc.expectedCode, MethodLabel: http.MethodPost}
			assert.InDelta(t, 0.0, testutil.ToFloat64(ci.inFlight), 0, "in-flight after transaction")
			assert.InDelta(t, 1.0, testutil.ToFloat64(ci.count.With(l)), 0)
			assert.InDelta(t, tc.expectedErrors, testutil.ToFloat64(ci.errorCount.With(l)), 0)

			count, sum := observed(t, ci.duration, l)
			assert.Equal(t, uint64(1), count)
			assert.InDelta(t, 100.0, sum, 0, "duration in milliseconds")

			count, sum = observed(t, ci.requestSize, l)
			assert.Equal(t, uint64(1), count)
			assert.InDelta(t, float64(len(body)), sum, 0, "request size in bytes")
		})
	}
}
