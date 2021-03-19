package touchhttp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/httpaux/httpmock"
	"github.com/xmidt-org/touchstone"
)

const (
	testServerName = "test"
	testClientName = "test"

	testDuration time.Duration = 151 * time.Millisecond
)

// BundleTestSuite is all the common functionality for testing metric bundles.
type BundleTestSuite struct {
	suite.Suite

	// now is a known start time for all clocks
	now time.Time
}

var _ suite.SetupTestSuite = (*BundleTestSuite)(nil)

func (suite *BundleTestSuite) SetupTest() {
	suite.now = time.Now()
}

func (suite *BundleTestSuite) Printf(format string, args ...interface{}) {
	suite.T().Logf(format, args...)
}

// newRegistry constructs a registry and its wrapping factory with a prebuilt
// environment for testing.
func (suite *BundleTestSuite) newRegistry() (*prometheus.Registry, *touchstone.Factory) {
	r := prometheus.NewPedanticRegistry()
	f := touchstone.NewFactory(touchstone.Config{}, suite, r)
	return r, f
}

// encode writes a gatherer's state to the writer using the testutil package.
// This enables comparing the state of two registries.
func (suite *BundleTestSuite) encode(o io.Writer, g prometheus.Gatherer) {
	raw, err := g.Gather()
	suite.Require().NoError(err)

	encoder := expfmt.NewEncoder(o, expfmt.FmtText)
	if closer, ok := encoder.(expfmt.Closer); ok {
		defer closer.Close()
	}

	for _, mf := range raw {
		suite.Require().NoError(encoder.Encode(mf))
	}
}

// newClock creates a now implementation that produces a known duration.
func (suite *BundleTestSuite) newClock() clock {
	return clock{
		start:    suite.now,
		duration: testDuration,
	}
}

type ServerBundleTestSuite struct {
	BundleTestSuite

	// expectedInFlight is the expected flat text output of a gatherer
	// while a request is in flight.  We can pregenerate this because
	// it's always the same for any server.  Only the inflight gauge
	// is incremented when the handler is executing.
	expectedInFlight []byte

	// expectedStatusCode is what handler will write and what assertMetrics
	// will verify.  If 0, no status code from handler is expected.
	expectedStatusCode int

	// registry is the current test's registry
	registry *prometheus.Registry

	// bundle is the current test's bundle
	bundle ServerBundle
}

func (suite *ServerBundleTestSuite) SetupTest() {
	suite.BundleTestSuite.SetupTest()
	registry, expected := suite.newServerBundle()
	expected.inFlight.With(prometheus.Labels{ServerLabel: testServerName}).Inc()

	var output bytes.Buffer
	suite.encode(&output, registry)
	suite.expectedInFlight = output.Bytes()
}

// newServerBundle returns a ServerBundle with the registry that backs it.  Useful
// for both test instances and for use with prometheus' testutil package.
func (suite *ServerBundleTestSuite) newServerBundle() (*prometheus.Registry, ServerBundle) {
	r, f := suite.newRegistry()
	c := suite.newClock()
	sb, err := NewServerBundle(f, c.Now)
	suite.Require().NoError(err)
	suite.Nil(sb.errorCount)
	return r, sb
}

// setupServerBundle sets up the suite state for a handler run and
// metrics assertions.
func (suite *ServerBundleTestSuite) setupServerBundle(expectedStatusCode int) (*prometheus.Registry, ServerBundle) {
	suite.expectedStatusCode = expectedStatusCode
	r, b := suite.newServerBundle()
	suite.registry = r
	suite.bundle = b
	return r, b
}

// assertMetrics verifies the metrics for all tests.
func (suite *ServerBundleTestSuite) assertMetrics(r *http.Request) {
	statusCode := suite.expectedStatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	labels := prometheus.Labels{
		ServerLabel: testServerName,
		MethodLabel: cleanMethod(r.Method),
		CodeLabel:   strconv.Itoa(statusCode),
	}

	expectedRegistry, expected := suite.newServerBundle()
	expected.inFlight.With(prometheus.Labels{ServerLabel: testServerName}).Inc()
	expected.inFlight.With(prometheus.Labels{ServerLabel: testServerName}).Dec()
	expected.counter.With(labels).Inc()
	expected.requestSize.With(labels).Observe(float64(r.ContentLength))
	expected.duration.With(labels).Observe(float64(testDuration / time.Millisecond))

	var expectedText bytes.Buffer
	suite.encode(&expectedText, expectedRegistry)
	suite.NoError(
		testutil.GatherAndCompare(suite.registry, &expectedText),
	)
}

// handler is a test HTTP handler function.  It verifies in flight metric state and writes
// the expectedStatusCode if nonzero.
func (suite *ServerBundleTestSuite) handler(rw http.ResponseWriter, r *http.Request) {
	var output bytes.Buffer
	suite.encode(&output, suite.registry)
	suite.Equal(suite.expectedInFlight, output.Bytes())

	if suite.expectedStatusCode > 0 {
		rw.WriteHeader(suite.expectedStatusCode)
	}
}

func (suite *ServerBundleTestSuite) TestDefaultNow() {
	r := prometheus.NewRegistry()
	f := touchstone.NewFactory(touchstone.Config{}, nil, r)
	sb, err := NewServerBundle(f, nil)
	suite.Require().NoError(err)
	suite.Require().NotNil(sb.now)

	t := sb.now()
	suite.False(t.IsZero())
}

func (suite *ServerBundleTestSuite) TestUsage() {
	testCases := []struct {
		method             string
		body               io.ReadCloser
		expectedStatusCode int
	}{
		{
			method:             "GET",
			expectedStatusCode: 299,
		},
		{
			method:             "POST",
			body:               httpmock.BodyString("test body text"),
			expectedStatusCode: 0,
		},
		{
			method:             "THISISNOTAMETHOD",
			expectedStatusCode: 218,
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			var (
				_, actual = suite.setupServerBundle(testCase.expectedStatusCode)
				decorated = actual.ForServer(testServerName).Then(http.HandlerFunc(suite.handler))
				response  = httptest.NewRecorder()

				// create the request this way so we can handle unsupported methods
				request = httptest.NewRequest("GET", "/test", testCase.body)
			)

			request.Method = testCase.method // allow for unrecognized methods

			decorated.ServeHTTP(response, request)
			suite.assertMetrics(request)
		})
	}
}

func TestServerBundle(t *testing.T) {
	suite.Run(t, new(ServerBundleTestSuite))
}

type ClientBundleTestSuite struct {
	BundleTestSuite

	// expectedInFlight is the expected flat text output of a gatherer
	// while a request is in flight.  We can pregenerate this because
	// it's always the same for any client.  Only the inflight gauge
	// is incremented when the handler is executing.
	expectedInFlight []byte

	// registry is the current test's registry
	registry *prometheus.Registry

	// bundle is the current test's bundle
	bundle ClientBundle
}

func (suite *ClientBundleTestSuite) SetupTest() {
	suite.BundleTestSuite.SetupTest()
	registry, expected := suite.newClientBundle()
	expected.inFlight.With(prometheus.Labels{ClientLabel: testClientName}).Inc()

	var output bytes.Buffer
	suite.encode(&output, registry)
	suite.expectedInFlight = output.Bytes()
}

// assertInFlight is a mock Run function that verifies in flight metric state.
func (suite *ClientBundleTestSuite) assertInFlight(arguments mock.Arguments) {
	var output bytes.Buffer
	suite.encode(&output, suite.registry)
	suite.Equal(suite.expectedInFlight, output.Bytes())
}

// newClientBundle returns a ServerBundle with the registry that backs it.  Useful
// for both test instances and for use with prometheus' testutil package.
func (suite *ClientBundleTestSuite) newClientBundle() (*prometheus.Registry, ClientBundle) {
	r, f := suite.newRegistry()
	c := suite.newClock()
	cb, err := NewClientBundle(f, c.Now)
	suite.Require().NoError(err)
	suite.Require().NotNil(cb.errorCount)
	return r, cb
}

// setupClientBundle sets up the suite state for round trip run with
// metrics assertions.
func (suite *ClientBundleTestSuite) setupClientBundle() (*prometheus.Registry, ClientBundle) {
	r, b := suite.newClientBundle()
	suite.registry = r
	suite.bundle = b
	return r, b
}

// newRequest sets up a client request with a possibly malformed method.
func (suite *ClientBundleTestSuite) newRequest(method string, body io.ReadCloser) *http.Request {
	r, err := http.NewRequest("GET", "/test", body)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	r.Method = method // support methods that http.NewRequest would fail on
	return r
}

// assertMetrics verifies the metrics for all tests.
func (suite *ClientBundleTestSuite) assertMetrics(request *http.Request, response *http.Response, err error) {
	statusCode := -1
	if response != nil {
		statusCode = response.StatusCode
	}

	labels := prometheus.Labels{
		ClientLabel: testClientName,
		MethodLabel: cleanMethod(request.Method),
		CodeLabel:   strconv.Itoa(statusCode),
	}

	expectedRegistry, expected := suite.newClientBundle()
	expected.inFlight.With(prometheus.Labels{ClientLabel: testClientName}).Inc()
	expected.inFlight.With(prometheus.Labels{ClientLabel: testClientName}).Dec()
	expected.counter.With(labels).Inc()
	expected.requestSize.With(labels).Observe(float64(request.ContentLength))
	expected.duration.With(labels).Observe(float64(testDuration / time.Millisecond))

	// this is the extra bit for client metrics
	if err != nil {
		expected.errorCount.With(labels).Inc()
	}

	var expectedText bytes.Buffer
	suite.encode(&expectedText, expectedRegistry)
	suite.NoError(
		testutil.GatherAndCompare(suite.registry, &expectedText),
	)
}
func (suite *ClientBundleTestSuite) TestDefaultNow() {
	r := prometheus.NewRegistry()
	f := touchstone.NewFactory(touchstone.Config{}, nil, r)
	cb, err := NewClientBundle(f, nil)
	suite.Require().NoError(err)
	suite.Require().NotNil(cb.now)

	t := cb.now()
	suite.False(t.IsZero())
}

func (suite *ClientBundleTestSuite) TestUsage() {
	testCases := []struct {
		method   string
		body     io.ReadCloser
		response *http.Response
		err      error
	}{
		{
			method: "GET",
			response: &http.Response{
				StatusCode: 299,
			},
		},
		{
			method: "GET",
			err:    errors.New("expected"),
		},
		{
			method: "WOTISTHIS",
			response: &http.Response{
				StatusCode: 348,
			},
		},
		{
			method: "WOTISTHIS",
			err:    errors.New("expected"),
		},
		{
			method: "POST",
			body:   httpmock.BodyString("testy mctest body"),
			response: &http.Response{
				StatusCode: 217,
			},
		},
		{
			method: "POST",
			body:   httpmock.BodyString("testy mctest body"),
			err:    errors.New("expected"),
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			var (
				_, actual    = suite.setupClientBundle()
				roundTripper = httpmock.NewRoundTripperSuite(suite)
				decorated    = actual.ForClient(testClientName).Then(&http.Client{Transport: roundTripper})
				request      = suite.newRequest(testCase.method, testCase.body)
			)

			roundTripper.OnAny().Return(testCase.response, testCase.err).Run(suite.assertInFlight).Once()
			response, err := decorated.Do(request)
			suite.True(response == testCase.response)
			suite.True(errors.Is(err, testCase.err))
			suite.assertMetrics(request, response, err)
			roundTripper.AssertExpectations()
		})
	}
}

func TestClientBundle(t *testing.T) {
	suite.Run(t, new(ClientBundleTestSuite))
}
