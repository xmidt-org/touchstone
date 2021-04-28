package touchhttp

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/multierr"
)

const (
	// DefaultServerCount is the default name of the counter that tracks the
	// total number of received server requests.
	DefaultServerCount = "server_request_count"

	// DefaultServerDuration is the default name of the observer that tracks
	// the total time taken by handlers to process requests.
	DefaultServerDuration = "server_request_duration_ms"

	// DefaultServerInFlight is the default name of the gauge that tracks the
	// instantaneous view of how many requests the handler is currently serving.
	DefaultServerInFlight = "server_requests_in_flight"

	// DefaultServerRequestSize is the default name of the observer that tracks the sizes
	// of requests received by handlers.
	DefaultServerRequestSize = "server_request_size"

	// DefaultClientCount is the default name of the counter that tracks the
	// total number of outgoing server requests.
	DefaultClientCount = "client_request_count"

	// DefaultClientDuration is the default name of the observer that tracks
	// the total time taken to send a request and receive a response.
	DefaultClientDuration = "client_request_duration_ms"

	// DefaultClientInFlight is the default name of the gauge that tracks the
	// instantaneous view of how many requests the client has currently pending.
	DefaultClientInFlight = "client_requests_in_flight"

	// DefaultClientRequestSize is the default name of the observer that tracks the sizes
	// of requests sent to servers.
	DefaultClientRequestSize = "client_request_size"

	// DefaultClientErrorCount is the default name of the count of total number of errors
	// (nil responses) that occurred since startup.
	DefaultClientErrorCount = "client_error_count"
)

var (
	// ErrReservedLabelName indicates that labels supplied to build an instrumenter
	// had one or more reserved label names.
	ErrReservedLabelName = fmt.Errorf(
		"%s and %s are reserved label names and are supplied automatically",
		CodeLabel,
		MethodLabel,
	)

	defaultServerCount = prometheus.CounterOpts{
		Name: DefaultServerCount,
		Help: "the total number of requests received since startup",
	}

	defaultServerInFlight = prometheus.GaugeOpts{
		Name: DefaultServerInFlight,
		Help: "the instantaneous number of requests currently being handled",
	}

	defaultServerDuration = prometheus.HistogramOpts{
		Name:    DefaultServerDuration,
		Help:    "the request duration in milliseconds",
		Buckets: []float64{62.5, 125, 250, 500, 1000, 5000, 10000, 20000, 40000, 80000, 160000},
	}

	defaultServerRequestSize = prometheus.HistogramOpts{
		Name: DefaultServerRequestSize,
		Help: "the size of handled requests in bytes",
		// TODO: add default buckets?
	}

	defaultClientCount = prometheus.CounterOpts{
		Name: DefaultClientCount,
		Help: "the total number of requests sent since startup",
	}

	defaultClientInFlight = prometheus.GaugeOpts{
		Name: DefaultClientInFlight,
		Help: "the instantaneous number of requests currently pending",
	}

	defaultClientDuration = prometheus.HistogramOpts{
		Name:    DefaultClientDuration,
		Help:    "the total time, in milliseconds, between sending a request and receiving a response",
		Buckets: []float64{62.5, 125, 250, 500, 1000, 5000, 10000, 20000, 40000, 80000, 160000},
	}

	defaultClientRequestSize = prometheus.HistogramOpts{
		Name: DefaultClientRequestSize,
		Help: "the size of outgoing requests in bytes",
		// TODO: add default buckets?
	}

	defaultClientErrorCount = prometheus.CounterOpts{
		Name: DefaultClientErrorCount,
		Help: "the total number of errors (nil responses) since startup",
	}
)

// labelNames extracts the names from the given Labels map.  This function
// performs a single allocation and returns two slices of names:  (1) extra
// are the label names other than the reserved names that are curried for
// a given server or client, and (2) full are all the label names, including
// the reserved names.
func labelNames(l prometheus.Labels) (extra, full []string, err error) {
	full = make([]string, 0, len(l))
	for n := range l {
		if n == CodeLabel || n == MethodLabel {
			err = ErrReservedLabelName
			return
		} else {
			full = append(full, n)
		}
	}

	full = append(full, CodeLabel, MethodLabel)
	extra = full[0 : len(full)-2]
	return
}

func newCounterVec(f *touchstone.Factory, o prometheus.CounterOpts, labelNames []string, curry prometheus.Labels) (*prometheus.CounterVec, error) {
	cv, err := f.NewCounterVec(o, labelNames...)
	err = touchstone.ExistingCollector(&cv, err)
	if err == nil {
		cv, err = cv.CurryWith(curry)
	}

	return cv, err
}

func newGauge(f *touchstone.Factory, o prometheus.GaugeOpts, labelNames []string, curry prometheus.Labels) (prometheus.Gauge, error) {
	gv, err := f.NewGaugeVec(o, labelNames...)
	err = touchstone.ExistingCollector(&gv, err)
	if err == nil {
		return gv.GetMetricWith(curry)
	}

	return nil, err
}

func newObserverVec(f *touchstone.Factory, o interface{}, labelNames []string, curry prometheus.Labels) (prometheus.ObserverVec, error) {
	ov, err := f.NewObserverVec(o, labelNames...)
	err = touchstone.ExistingCollector(&ov, err)
	if err == nil {
		ov, err = ov.CurryWith(curry)
	}

	return ov, err
}

type ServerBundle struct {
	// Count describes the options used for the total request counter
	Count prometheus.CounterOpts

	// InFlight describes the options used for the instaneous request gauge
	InFlight prometheus.GaugeOpts

	// RequestSize describes the options for the request size observer.  A panic
	// will result if this field is not nil, a prometheus.HistogramOpts, or
	// a prometheus.SummaryOpts.
	RequestSize interface{}

	// Duration describes the options for the request duration observer.  A panic
	// will result if this field is not nil, a prometheus.HistogramOpts, or
	// a prometheus.SummaryOpts.
	Duration interface{}

	// Now is the strategy for extracting the current system time.  If unset,
	// time.Now is used.
	Now func() time.Time
}

// ForServer produces a ServerInstrumenter using this bundle of metric options.  The internal
// server defaults are applied to each option for fields that are unset, e.g. metric names.
//
// This method may be called multiple times with different prometheus.Labels.  In that case
// each call must supply the same label names, or the underlying prometheus library will
// return an error.
func (sb ServerBundle) ForServer(f *touchstone.Factory, l prometheus.Labels) (si ServerInstrumenter, err error) {
	var extra, full []string
	extra, full, err = labelNames(l)
	if err != nil {
		return
	}

	si.now = sb.Now
	if si.now == nil {
		si.now = time.Now
	}

	var metricErr error

	touchstone.ApplyDefaults(&sb.Count, defaultServerCount)
	si.count, metricErr = newCounterVec(f, sb.Count, full, l)
	multierr.AppendInto(&err, metricErr)

	// InFlight is slightly different, as it doesn't have code or method labels
	touchstone.ApplyDefaults(&sb.InFlight, defaultServerInFlight)
	si.inFlight, metricErr = newGauge(f, sb.InFlight, extra, l)
	multierr.AppendInto(&err, metricErr)

	touchstone.ApplyDefaults(&sb.RequestSize, defaultServerRequestSize)
	si.requestSize, metricErr = newObserverVec(f, sb.RequestSize, full, l)
	multierr.AppendInto(&err, metricErr)

	touchstone.ApplyDefaults(&sb.Duration, defaultServerDuration)
	si.duration, metricErr = newObserverVec(f, sb.Duration, full, l)
	multierr.AppendInto(&err, metricErr)

	return
}

type ClientBundle struct {
	// Count describes the options used for the total request counter
	Count prometheus.CounterOpts

	// InFlight describes the options used for the instaneous request gauge
	InFlight prometheus.GaugeOpts

	// RequestSize describes the options for the request size observer.  A panic
	// will result if this field is not nil, a prometheus.HistogramOpts, or
	// a prometheus.SummaryOpts.
	RequestSize interface{}

	// Duration describes the options for the request duration observer.  A panic
	// will result if this field is not either a prometheus.HistogramOpts or a prometheus.SummaryOpts.
	Duration interface{}

	// ErrorCount describes the options for the error counter.
	ErrorCount prometheus.CounterOpts

	// Now is the strategy for extracting the current system time.  If unset,
	// time.Now is used.
	Now func() time.Time
}

// ForClient produces a ClientInstrumenter using this bundle of metric options.  The internal
// server defaults are applied to each option for fields that are unset, e.g. metric names.
//
// This method may be called multiple times with different prometheus.Labels.  In that case
// each call must supply the same label names, or the underlying prometheus library will
// return an error.
func (cb ClientBundle) ForClient(f *touchstone.Factory, l prometheus.Labels) (ci ClientInstrumenter, err error) {
	var extra, full []string
	extra, full, err = labelNames(l)
	if err != nil {
		return
	}

	ci.now = cb.Now
	if ci.now == nil {
		ci.now = time.Now
	}

	var metricErr error

	touchstone.ApplyDefaults(&cb.Count, defaultClientCount)
	ci.count, metricErr = newCounterVec(f, cb.Count, full, l)
	multierr.AppendInto(&err, metricErr)

	// InFlight is slightly different, as it doesn't have code or method labels
	touchstone.ApplyDefaults(&cb.InFlight, defaultClientInFlight)
	ci.inFlight, metricErr = newGauge(f, cb.InFlight, extra, l)
	multierr.AppendInto(&err, metricErr)

	touchstone.ApplyDefaults(&cb.RequestSize, defaultClientRequestSize)
	ci.requestSize, metricErr = newObserverVec(f, cb.RequestSize, full, l)
	multierr.AppendInto(&err, metricErr)

	touchstone.ApplyDefaults(&cb.Duration, defaultClientDuration)
	ci.duration, metricErr = newObserverVec(f, cb.Duration, full, l)
	multierr.AppendInto(&err, metricErr)

	touchstone.ApplyDefaults(&cb.ErrorCount, defaultClientErrorCount)
	ci.errorCount, metricErr = newCounterVec(f, cb.ErrorCount, full, l)
	multierr.AppendInto(&err, metricErr)

	return
}
