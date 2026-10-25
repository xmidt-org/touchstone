// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		lvs           []string
		expectedNames []string
		expectedCurry prometheus.Labels
		expectedErr   error
	}{
		{
			name:          "Empty",
			expectedNames: []string{},
			expectedCurry: prometheus.Labels{},
		},
		{
			name:          "Pairs",
			lvs:           []string{"a", "1", "b", "2"},
			expectedNames: []string{"a", "b"},
			expectedCurry: prometheus.Labels{"a": "1", "b": "2"},
		},
		{
			name:        "OddCount",
			lvs:         []string{"a"},
			expectedErr: ErrInvalidLabelCount,
		},
		{
			name:        "ReservedCode",
			lvs:         []string{CodeLabel, "200"},
			expectedErr: ErrReservedLabelName,
		},
		{
			name:        "ReservedMethod",
			lvs:         []string{"a", "1", MethodLabel, "GET"},
			expectedErr: ErrReservedLabelName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			names, curry, err := labelNames(tc.lvs)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedNames, names)
			assert.Equal(t, tc.expectedCurry, curry)
		})
	}
}

func TestServerBundleObserverOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		bundle    ServerBundle
		expectErr bool
	}{
		{
			name: "Histograms",
			bundle: ServerBundle{
				RequestSize: prometheus.HistogramOpts{Name: "custom_size", Buckets: []float64{1, 10}},
				Duration:    prometheus.HistogramOpts{Name: "custom_duration"},
			},
		},
		{
			name: "Summaries",
			bundle: ServerBundle{
				RequestSize: prometheus.SummaryOpts{Name: "custom_size"},
				Duration:    prometheus.SummaryOpts{Name: "custom_duration"},
			},
		},
		{
			name:      "InvalidRequestSize",
			bundle:    ServerBundle{RequestSize: "not an opts struct"},
			expectErr: true,
		},
		{
			name:      "InvalidDuration",
			bundle:    ServerBundle{Duration: 123},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			si, err := tc.bundle.NewInstrumenter()(newTestFactory(t))
			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, si.requestSize)
			assert.NotNil(t, si.duration)
		})
	}
}

func TestClientBundleObserverOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		bundle    ClientBundle
		expectErr bool
	}{
		{
			name: "Histograms",
			bundle: ClientBundle{
				RequestSize: prometheus.HistogramOpts{Name: "custom_size", Buckets: []float64{1, 10}},
				Duration:    prometheus.HistogramOpts{Name: "custom_duration"},
			},
		},
		{
			name: "Summaries",
			bundle: ClientBundle{
				RequestSize: prometheus.SummaryOpts{Name: "custom_size"},
				Duration:    prometheus.SummaryOpts{Name: "custom_duration"},
			},
		},
		{
			name:      "InvalidRequestSize",
			bundle:    ClientBundle{RequestSize: "not an opts struct"},
			expectErr: true,
		},
		{
			name:      "InvalidDuration",
			bundle:    ClientBundle{Duration: 123},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ci, err := tc.bundle.NewInstrumenter()(newTestFactory(t))
			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, ci.requestSize)
			assert.NotNil(t, ci.duration)
		})
	}
}

func TestNewInstrumenterErrors(t *testing.T) {
	t.Parallel()

	t.Run("InvalidLabels", func(t *testing.T) {
		t.Parallel()

		f := newTestFactory(t)

		_, err := ServerBundle{}.NewInstrumenter("odd")(f)
		assert.ErrorIs(t, err, ErrInvalidLabelCount)

		_, err = ClientBundle{}.NewInstrumenter(MethodLabel, "GET")(f)
		assert.ErrorIs(t, err, ErrReservedLabelName)
	})

	// Registering the same metric names a second time with a different set of
	// label names is rejected by prometheus, and that error must surface.
	t.Run("InconsistentLabelNames", func(t *testing.T) {
		t.Parallel()

		f := newTestFactory(t)

		_, err := ServerBundle{}.NewInstrumenter("a", "1")(f)
		require.NoError(t, err)
		_, err = ServerBundle{}.NewInstrumenter("b", "2")(f)
		assert.Error(t, err)

		_, err = ClientBundle{}.NewInstrumenter("a", "1")(f)
		require.NoError(t, err)
		_, err = ClientBundle{}.NewInstrumenter("b", "2")(f)
		assert.Error(t, err)
	})

	// The same label names with different values must share the underlying metrics.
	t.Run("ConsistentLabelNames", func(t *testing.T) {
		t.Parallel()

		f := newTestFactory(t)

		_, err := ServerBundle{}.NewInstrumenter(ServerLabel, "main")(f)
		require.NoError(t, err)
		_, err = ServerBundle{}.NewInstrumenter(ServerLabel, "health")(f)
		assert.NoError(t, err)
	})
}
