package touchhttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/httpaux"
	"github.com/xmidt-org/httpaux/client"
	"github.com/xmidt-org/httpaux/observe"
	"github.com/xmidt-org/touchstone"
)

const (
	// ServerLabel is the metric label containing the name of the HTTP server.
	ServerLabel = "server"

	// ClientLabel is the metric label containing the name of the HTTP client.
	ClientLabel = "client"

	// CodeLabel is the metric label containing the HTTP response code.
	CodeLabel = "code"

	// MethodLabel is the metric label containing the HTTP request's method.
	MethodLabel = "method"

	// MethodUnrecognized is used when an HTTP method is not one of the
	// standard methods, as enumerated in the net/http package.
	MethodUnrecognized = "UNRECOGNIZED"

	// MetricServerRequestCount is the name of the counter that tracks the
	// total number of received server requests.
	MetricServerRequestCount = "server_request_count"

	// MetricServerRequestDuration is the name of the histogram that tracks
	// the total time taken by handlers to process requests.
	MetricServerRequestDuration = "server_request_duration_ms"

	// MetricServerRequestsInFlight is the name of the gauge that tracks the
	// instantaneous view of how many requests the handler is currently serving.
	MetricServerRequestsInFlight = "server_requests_in_flight"

	// MetricServerRequestSize is the name of the histogram that tracks the sizes
	// of requests received by handlers.
	MetricServerRequestSize = "server_request_size"

	// MetricClientRequestCount is the name of the counter that tracks the
	// total number of outgoing server requests.
	MetricClientRequestCount = "client_request_count"

	// MetricClientRequestDuration is the name of the histogram that tracks
	// the total time taken to send a request and receive a response.
	MetricClientRequestDuration = "client_request_duration_ms"

	// MetricClientRequestsInFlight is the name of the gauge that tracks the
	// instantaneous view of how many requests the client has currently pending.
	MetricClientRequestsInFlight = "client_requests_in_flight"

	// MetricClientRequestSize is the name of the histogram that tracks the sizes
	// of requests sent to servers.
	MetricClientRequestSize = "client_request_size"

	// MetricClientErrorCount is the total number of errors (nil responses) that
	// occurred since startup.
	MetricClientErrorCount = "client_error_count"
)

// metricHelp holds the Help text for each bundled metric.
var metricHelp = map[string]string{
	MetricServerRequestCount:     "the total number of requests received since startup",
	MetricServerRequestDuration:  "the request duration in milliseconds",
	MetricServerRequestsInFlight: "the instantaneous number of requests currently being handled",
	MetricServerRequestSize:      "the size of handled requests in bytes",

	MetricClientRequestCount:     "the total number of requests sent since startup",
	MetricClientRequestDuration:  "the total time, in milliseconds, between sending a request and receiving a response",
	MetricClientRequestsInFlight: "the instantaneous number of requests currently pending",
	MetricClientRequestSize:      "the size of outgoing requests in bytes",
	MetricClientErrorCount:       "the total number of errors (nil responses) since startup",
}

// recognizedMethods is the set of HTTP methods that are defined by the spec(s).
// Any method not in this map gets recorded as MethodUnknown.
var recognizedMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodConnect: true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

func cleanMethod(v string) string {
	if recognizedMethods[v] {
		return v
	}

	return MethodUnrecognized
}

// transaction represents a completed HTTP transaction.
type transaction struct {
	start       time.Time
	code        int
	method      string
	err         error // that came from a client
	requestSize int64
}

// bundle is a prebaked set of HTTP metrics.  A bundle is curried
// to create an instrumenter.
type bundle struct {
	counter     *prometheus.CounterVec
	inFlight    *prometheus.GaugeVec
	requestSize prometheus.ObserverVec
	duration    prometheus.ObserverVec

	// only used in clients
	errorCount *prometheus.CounterVec

	now func() time.Time
}

// curry sets values for any labels that are constant for a server or client.
// For example, the server's name is curried away to produce an instrumenter
// specific to that server.
func (b bundle) curry(l prometheus.Labels) instrumenter {
	i := instrumenter{
		counter:     b.counter.MustCurryWith(l),
		inFlight:    b.inFlight.With(l),
		requestSize: b.requestSize.MustCurryWith(l),
		duration:    b.duration.MustCurryWith(l),
		now:         b.now,
	}

	if b.errorCount != nil {
		i.errorCount = b.errorCount.MustCurryWith(l)
	}

	return i
}

// instrumenter is the common logic that decorates HTTP transactions for
// both clients and servers.
type instrumenter struct {
	counter     *prometheus.CounterVec
	inFlight    prometheus.Gauge
	requestSize prometheus.ObserverVec
	duration    prometheus.ObserverVec

	// only used in clients
	errorCount *prometheus.CounterVec

	now func() time.Time
}

// begin records the start of an HTTP transaction
func (i instrumenter) begin(r *http.Request) transaction {
	i.inFlight.Inc()
	return transaction{
		start:       i.now(),
		method:      cleanMethod(r.Method),
		requestSize: r.ContentLength,
	}
}

func (i instrumenter) endHandle(sc observe.StatusCoder, t transaction) {
	t.code = sc.StatusCode()
	if t.code == 0 {
		// this can happen if a decorated handler never wrote a status code
		t.code = http.StatusOK
	}

	i.end(t)
}

func (i instrumenter) endDo(response *http.Response, err error, t transaction) {
	if response != nil {
		t.code = response.StatusCode
	} else {
		t.code = -1
	}

	t.err = err
	i.end(t)
}

// end records the end of an HTTP transaction
func (i instrumenter) end(t transaction) {
	i.inFlight.Dec()

	l := prometheus.Labels{
		MethodLabel: t.method,
		CodeLabel:   strconv.Itoa(t.code),
	}

	i.counter.With(l).Inc()
	elapsed := i.now().Sub(t.start)
	i.duration.With(l).Observe(
		float64(elapsed / time.Millisecond),
	)

	i.requestSize.With(l).Observe(
		float64(t.requestSize),
	)

	if i.errorCount != nil && t.err != nil {
		i.errorCount.With(l).Inc()
	}
}

// ServerBundle is a prebaked set of serverside HTTP metrics.  A ServerBundle
// is used to create middleware in the form of a ServerInstrumenter.
type ServerBundle struct {
	bundle
}

// NewServerBundle constructs a bundle of HTTP  metrics using the given Factory.
// This function should be called at most once for any given Factory, or a
// prometheus.AlreadyRegisteredError will occur.
//
// If now is nil, time.Now is used.
func NewServerBundle(f *touchstone.Factory, now func() time.Time) (sb ServerBundle, err error) {
	sb.now = now
	if sb.now == nil {
		sb.now = time.Now
	}

	sb.counter, err = f.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricServerRequestCount,
			Help: metricHelp[MetricServerRequestCount],
		}, ServerLabel, CodeLabel, MethodLabel,
	)

	if err == nil {
		sb.duration, err = f.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: MetricServerRequestDuration,
				Help: metricHelp[MetricServerRequestDuration],
			}, ServerLabel, CodeLabel, MethodLabel,
		)
	}

	if err == nil {
		// NOTE: the inflight gauge can't have code and method labels, because when
		// the gauge is incremented the decorated HTTP client or handler hasn't executed yet.
		sb.inFlight, err = f.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricServerRequestsInFlight,
				Help: metricHelp[MetricServerRequestsInFlight],
			}, ServerLabel,
		)
	}

	if err == nil {
		sb.requestSize, err = f.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: MetricServerRequestSize,
				Help: metricHelp[MetricServerRequestSize],
			}, ServerLabel, CodeLabel, MethodLabel,
		)
	}

	return
}

// ForServer curries this bundle with the given server name and produces
// a ServerInstrumenter.
func (sb ServerBundle) ForServer(server string) (si ServerInstrumenter) {
	si.instrumenter = sb.bundle.curry(prometheus.Labels{
		ServerLabel: server,
	})

	return
}

// ServerInstrumenter is a serverside middleware that provides http.Handler
// metrics.
type ServerInstrumenter struct {
	instrumenter
}

// Then is a server middleware that instruments the given handler.  This middleware
// is compatible with justinas/alice and gorilla/mux.
func (si ServerInstrumenter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w := observe.New(rw)
		t := si.begin(r)
		defer si.endHandle(w, t)

		next.ServeHTTP(w, r)
	})
}

// ClientBundle is a prebaked set of clientside HTTP metrics.  A ClientBundle
// is used to create middleware in the form of a ClientInstrumenter.
type ClientBundle struct {
	bundle
}

// NewClientBundle constructs a bundle of HTTP  metrics using the given Factory.
// This function should be called at most once for any given Factory, or a
// prometheus.AlreadyRegisteredError will occur.
//
// If now is nil, time.Now is used.
func NewClientBundle(f *touchstone.Factory, now func() time.Time) (cb ClientBundle, err error) {
	cb.now = now
	if cb.now == nil {
		cb.now = time.Now
	}

	cb.counter, err = f.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricClientRequestCount,
			Help: metricHelp[MetricClientRequestCount],
		}, ClientLabel, CodeLabel, MethodLabel,
	)

	if err == nil {
		cb.duration, err = f.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: MetricClientRequestDuration,
				Help: metricHelp[MetricClientRequestDuration],
			}, ClientLabel, CodeLabel, MethodLabel,
		)
	}

	if err == nil {
		// NOTE: the inflight gauge can't have code and method labels, because when
		// the gauge is incremented the decorated HTTP client or handler hasn't executed yet.
		cb.inFlight, err = f.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: MetricClientRequestsInFlight,
				Help: metricHelp[MetricClientRequestsInFlight],
			}, ClientLabel,
		)
	}

	if err == nil {
		cb.requestSize, err = f.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: MetricClientRequestSize,
				Help: metricHelp[MetricClientRequestSize],
			}, ClientLabel, CodeLabel, MethodLabel,
		)
	}

	if err == nil {
		cb.errorCount, err = f.NewCounterVec(
			prometheus.CounterOpts{
				Name: MetricClientErrorCount,
				Help: metricHelp[MetricClientErrorCount],
			}, ClientLabel, CodeLabel, MethodLabel,
		)
	}

	return
}

// ForClient curries this bundle with the given client name and produces
// a ClientInstrumenter.
func (cb ClientBundle) ForClient(client string) (ci ClientInstrumenter) {
	ci.instrumenter = cb.bundle.curry(prometheus.Labels{
		ClientLabel: client,
	})

	return
}

// ClientInstrumenter is a clientside middleware that provides HTTP client
// metrics.
type ClientInstrumenter struct {
	instrumenter
}

// Then is a client middleware that instruments the given client.  This middleware
// is compatible with httpaux.
func (ci ClientInstrumenter) Then(next httpaux.Client) httpaux.Client {
	return client.Func(func(request *http.Request) (response *http.Response, err error) {
		t := ci.begin(request)
		response, err = next.Do(request)
		ci.endDo(response, err, t)
		return
	})
}

var _ client.Constructor = ClientInstrumenter{}.Then
