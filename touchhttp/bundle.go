/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package touchhttp

import (
	"errors"
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

	// ErrInvalidLabelCount indicates that an odd number of name/value pairs were
	// passed when creating metrics.
	ErrInvalidLabelCount = errors.New("The number of label names and values must be even")

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

// labelNames takes a sequence of name/value pairs and converts that into
// a slice of names and a prometheus.Labels which should be used to curry
// the associated metric.
func labelNames(lvs []string) (names []string, curry prometheus.Labels, err error) {
	if len(lvs)%2 != 0 {
		err = ErrInvalidLabelCount
	}

	if err == nil {
		names = make([]string, 0, len(lvs)/2)
		curry = make(prometheus.Labels, len(names))
		for i, j := 0, 1; err == nil && i < len(lvs); i, j = i+2, j+2 {
			if lvs[i] == CodeLabel || lvs[i] == MethodLabel {
				err = ErrReservedLabelName
				continue
			}

			names = append(names, lvs[i])
			curry[lvs[i]] = lvs[j]
		}
	}

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

	// RequestSize describes the options for the request size observer.  If this
	// field is set, it must be either a prometheus.HistogramOpts or a prometheus.SummaryOpts.
	// The type of Opts struct will determine the type of metric created.
	RequestSize interface{}

	// Duration describes the options for the request duration observer.  If this field is
	// set, it must be either a prometheus.HistogramOpts or a prometheus.SummaryOpts.
	// The type of Opts struct will determine the type of metric created.
	Duration interface{}

	// Now is the strategy for extracting the current system time.  If unset,
	// time.Now is used.
	Now func() time.Time
}

func (sb ServerBundle) newRequestCount(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (*prometheus.CounterVec, error) {
	touchstone.ApplyDefaults(&sb.Count, defaultServerCount)
	return newCounterVec(f, sb.Count, labelNames, curry)
}

func (sb ServerBundle) newInFlight(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.Gauge, error) {
	touchstone.ApplyDefaults(&sb.InFlight, defaultServerInFlight)
	return newGauge(f, sb.InFlight, labelNames, curry)
}

func (sb ServerBundle) newRequestSize(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.ObserverVec, error) {
	var opts interface{}
	if sb.RequestSize != nil {
		switch t := sb.RequestSize.(type) {
		case prometheus.HistogramOpts:
			touchstone.ApplyDefaults(&t, defaultServerRequestSize)
			opts = t

		case prometheus.SummaryOpts:
			touchstone.ApplyDefaults(&t, defaultServerRequestSize)
			opts = t

		default:
			return nil, errors.New("ServerBundle.RequestSize must be nil, a prometheus.HistogramOpts, or a prometheus.SummaryOpts")
		}
	} else {
		clone := defaultServerRequestSize
		opts = clone
	}

	return newObserverVec(f, opts, labelNames, curry)
}

func (sb ServerBundle) newDuration(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.ObserverVec, error) {
	var opts interface{}
	if sb.Duration != nil {
		switch t := sb.Duration.(type) {
		case prometheus.HistogramOpts:
			touchstone.ApplyDefaults(&t, defaultServerDuration)
			opts = t

		case prometheus.SummaryOpts:
			touchstone.ApplyDefaults(&t, defaultServerDuration)
			opts = t

		default:
			return nil, errors.New("ServerBundle.Duration must be nil, a prometheus.HistogramOpts, or a prometheus.SummaryOpts")
		}
	} else {
		clone := defaultServerDuration
		opts = clone
	}

	return newObserverVec(f, opts, labelNames, curry)
}

// NewInstrumenter creates a constructor that can be passed to fx.Provide or annotated
// as needed.
//
// The namesAndValues are any extra, curried labels to apply to all the created
// metrics.  If multiple calls to this method on the same ServerBundle instance are made,
// the extra label names must match though the values may differ.
//
// If namesAndValues contatins and odd number of entries or if it contains any of
// the reserved label names used by this package, and error is returned by the returned
// constructor.
//
// Typical usage:
//
//   app := fx.New(
//     touchstone.Provide(), // bootstraps the metrics environment
//
//     fx.Provide(
//       // Create a single, unnamed ServerInstrumenter with no extra labels
//       touchhttp.ServerBundle{}.NewInstrumenter(),
//
//       // Create a named ServerInstrumenter with a label identifying a particular server
//       fx.Annotated{
//         Name: "servers.main",
//         Target: touchhttp.ServerBundle{}.NewInstrumenter(
//           touchhttp.ServerLabel, "servers.main",
//         ),
//       },
//     ),
//   )
func (sb ServerBundle) NewInstrumenter(namesAndValues ...string) func(*touchstone.Factory) (ServerInstrumenter, error) {
	return func(f *touchstone.Factory) (si ServerInstrumenter, err error) {
		var (
			extraNames []string
			curry      prometheus.Labels
		)

		extraNames, curry, err = labelNames(namesAndValues)
		if err != nil {
			return
		}

		// fullNames will include the extra names plus code and method labels
		fullNames := make([]string, 0, len(extraNames)+2)
		fullNames = append(fullNames, extraNames...)
		fullNames = append(fullNames, CodeLabel, MethodLabel)

		si.now = sb.Now
		if si.now == nil {
			si.now = time.Now
		}

		var metricErr error

		si.count, metricErr = sb.newRequestCount(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		// InFlight is slightly different, as it doesn't have code or method labels
		si.inFlight, metricErr = sb.newInFlight(f, extraNames, curry)
		multierr.AppendInto(&err, metricErr)

		si.requestSize, metricErr = sb.newRequestSize(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		si.duration, metricErr = sb.newDuration(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		return
	}
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

func (cb ClientBundle) newRequestCount(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (*prometheus.CounterVec, error) {
	touchstone.ApplyDefaults(&cb.Count, defaultClientCount)
	return newCounterVec(f, cb.Count, labelNames, curry)
}

func (cb ClientBundle) newInFlight(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.Gauge, error) {
	touchstone.ApplyDefaults(&cb.InFlight, defaultClientInFlight)
	return newGauge(f, cb.InFlight, labelNames, curry)
}

func (cb ClientBundle) newRequestSize(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.ObserverVec, error) {
	var opts interface{}
	if cb.RequestSize != nil {
		switch t := cb.RequestSize.(type) {
		case prometheus.HistogramOpts:
			touchstone.ApplyDefaults(&t, defaultClientRequestSize)
			opts = t

		case prometheus.SummaryOpts:
			touchstone.ApplyDefaults(&t, defaultClientRequestSize)
			opts = t

		default:
			return nil, errors.New("ClientBundle.RequestSize must be nil, a prometheus.HistogramOpts, or a prometheus.SummaryOpts")
		}
	} else {
		clone := defaultClientRequestSize
		opts = clone
	}

	return newObserverVec(f, opts, labelNames, curry)
}

func (cb ClientBundle) newDuration(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (prometheus.ObserverVec, error) {
	var opts interface{}
	if cb.Duration != nil {
		switch t := cb.Duration.(type) {
		case prometheus.HistogramOpts:
			touchstone.ApplyDefaults(&t, defaultClientDuration)
			opts = t

		case prometheus.SummaryOpts:
			touchstone.ApplyDefaults(&t, defaultClientDuration)
			opts = t

		default:
			return nil, errors.New("ClientBundle.Duration must be nil, a prometheus.HistogramOpts, or a prometheus.SummaryOpts")
		}
	} else {
		clone := defaultClientDuration
		opts = clone
	}

	return newObserverVec(f, opts, labelNames, curry)
}

func (cb ClientBundle) newErrorCount(f *touchstone.Factory, labelNames []string, curry prometheus.Labels) (*prometheus.CounterVec, error) {
	touchstone.ApplyDefaults(&cb.ErrorCount, defaultClientErrorCount)
	return newCounterVec(f, cb.ErrorCount, labelNames, curry)
}

// NewInstrumenter creates a constructor that can be passed to fx.Provide.  The returned constructor
// creates a ClientInstrumenter given a *touchstone.Factory.
//
// Similar typical usage to ServerBundle.NewInstrumenter:
//
//   app := fx.New(
//     touchstone.Provide(), // bootstraps the metrics environment
//
//     fx.Provide(
//       // Create a single, unnamed ClientInstrumenter with no extra labels
//       touchhttp.ClientBundle{}.NewInstrumenter(),
//
//       // Create a named ClientInstrumenter with a label identifying a particular client
//       fx.Annotated{
//         Name: "clients.main",
//         Target: touchhttp.ClientBundle{}.NewInstrumenter(
//           touchhttp.ClientLabel, "clients.main",
//         ),
//       },
//     ),
//   )
func (cb ClientBundle) NewInstrumenter(namesAndValues ...string) func(*touchstone.Factory) (ClientInstrumenter, error) {
	return func(f *touchstone.Factory) (ci ClientInstrumenter, err error) {
		var (
			extraNames []string
			curry      prometheus.Labels
		)

		extraNames, curry, err = labelNames(namesAndValues)
		if err != nil {
			return
		}

		// fullNames will include the extra names plus code and method labels
		fullNames := make([]string, 0, len(extraNames)+2)
		fullNames = append(fullNames, extraNames...)
		fullNames = append(fullNames, CodeLabel, MethodLabel)

		ci.now = cb.Now
		if ci.now == nil {
			ci.now = time.Now
		}

		var metricErr error

		ci.count, metricErr = cb.newRequestCount(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		// InFlight is slightly different, as it doesn't have code or method labels
		ci.inFlight, metricErr = cb.newInFlight(f, extraNames, curry)
		multierr.AppendInto(&err, metricErr)

		ci.requestSize, metricErr = cb.newRequestSize(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		ci.duration, metricErr = cb.newDuration(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		ci.errorCount, metricErr = cb.newErrorCount(f, fullNames, curry)
		multierr.AppendInto(&err, metricErr)

		return
	}
}
