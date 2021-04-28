package touchhttp

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/httpaux"
	"github.com/xmidt-org/httpaux/client"
	"github.com/xmidt-org/httpaux/observe"
)

// transaction represents a completed HTTP transaction.
type transaction struct {
	start       time.Time
	code        int
	method      string
	err         error // that came from a client
	requestSize int64
}

// instrumenter is the common logic that decorates HTTP transactions for
// both clients and servers.
type instrumenter struct {
	count       *prometheus.CounterVec
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
		method:      r.Method,
		requestSize: r.ContentLength,
	}
}

func (i instrumenter) endHandle(sc observe.StatusCoder, t transaction) {
	t.code = sc.StatusCode()
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

	l := prometheus.Labels(NewLabels(t.code, t.method))

	i.count.With(l).Inc()
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
