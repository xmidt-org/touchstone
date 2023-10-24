// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	// ErrNoMetricName indicates that a prometheus *Opts struct did not set the Name field.
	ErrNoMetricName = errors.New("A metric Name is required")
)

// Factory handles creation and registration of metrics.
//
// This type serves a similar purpose to the promauto package.  Instead of registering
// metrics with a global singleton, it uses the injected prometheus.Registerer.
// In addition, any DefaultNamespace and DefaultSubsystem set on the Config object
// are enforced for every metric created through the Factory instance.
//
// If a *zap.Logger is supplied, it is used to log warnings about missing Help
// in *Opts structs.
//
// This package's functions that match metric types, e.g. Counter, CounterVec, etc, use
// a Factory instance injected from the enclosing fx.App.  Those functions are generally
// preferred to using a Factory directly, since they emit their metrics as components which
// can then be injected as needed.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto
type Factory struct {
	defaults   prometheus.Opts
	logger     *zap.Logger
	registerer prometheus.Registerer
}

// NewFactory produces a Factory that uses the supplied registry.
func NewFactory(cfg Config, l *zap.Logger, r prometheus.Registerer) *Factory {
	return &Factory{
		defaults: prometheus.Opts{
			Namespace: cfg.DefaultNamespace,
			Subsystem: cfg.DefaultSubsystem,
		},
		logger:     l,
		registerer: r,
	}
}

func (f *Factory) checkName(v string) error {
	if len(v) == 0 {
		return ErrNoMetricName
	}

	return nil
}

func (f *Factory) warnOnNoHelp(name, help string) {
	if len(help) == 0 && f.logger != nil {
		f.logger.Warn("No help set for metric", zap.String("name", name))
	}
}

// DefaultNamespace returns the namespace used to register metrics
// when no Namespace is specified in the *Opts struct.  This may be
// empty to indicate that there is no default.
func (f *Factory) DefaultNamespace() string {
	return f.defaults.Namespace
}

// DefaultSubsystem returns the subsystem used to register metrics
// when no Subsystem is specified in the *Opts struct.  This may be
// empty to indicate that there is no default.
func (f *Factory) DefaultSubsystem() string {
	return f.defaults.Subsystem
}

// New creates a dynamically typed metric based on the concrete type passed as options.
// For example, if passed a prometheus.CounterOpts, this method creates and registers
// a prometheus.Counter.
//
// The o parameter must be one of the following, or this method panics:
//
//   - prometheus.CounterOpts
//   - *prometheus.CounterOpts
//   - prometheus.GaugeOpts
//   - *prometheus.GaugeOpts
//   - prometheus.HistogramOpts
//   - *prometheus.HistogramOpts
//   - prometheus.SummaryOpts
//   - *prometheus.SummaryOpts
func (f *Factory) New(o interface{}) (m prometheus.Collector, err error) {
	switch opts := o.(type) {
	case prometheus.CounterOpts:
		m, err = f.NewCounter(opts)

	case *prometheus.CounterOpts:
		m, err = f.NewCounter(*opts)

	case prometheus.GaugeOpts:
		m, err = f.NewGauge(opts)

	case *prometheus.GaugeOpts:
		m, err = f.NewGauge(*opts)

	case prometheus.HistogramOpts:
		var obs prometheus.Observer
		obs, err = f.NewHistogram(opts)
		if err == nil {
			m = obs.(prometheus.Collector)
		}

	case *prometheus.HistogramOpts:
		var obs prometheus.Observer
		obs, err = f.NewHistogram(*opts)
		if err == nil {
			m = obs.(prometheus.Collector)
		}

	case prometheus.SummaryOpts:
		var obs prometheus.Observer
		obs, err = f.NewSummary(opts)
		if err == nil {
			m = obs.(prometheus.Collector)
		}

	case *prometheus.SummaryOpts:
		var obs prometheus.Observer
		obs, err = f.NewSummary(*opts)
		if err == nil {
			m = obs.(prometheus.Collector)
		}

	default:
		panic(fmt.Errorf("%T is not a recognized prometheus xxxOpts struct", o))
	}

	return
}

// NewVec creates a dynamically typed metric vector based on the concrete type passed as options.
// For example, if passed a prometheus.CounterOpts, this method creates and registers
// a *prometheus.CounterVec.
//
// The o parameter must be one of the following, or this method panics:
//
//   - prometheus.CounterOpts
//   - *prometheus.CounterOpts
//   - prometheus.GaugeOpts
//   - *prometheus.GaugeOpts
//   - prometheus.HistogramOpts
//   - *prometheus.HistogramOpts
//   - prometheus.SummaryOpts
//   - *prometheus.SummaryOpts
func (f *Factory) NewVec(o interface{}, labelNames ...string) (m prometheus.Collector, err error) {
	switch opts := o.(type) {
	case prometheus.CounterOpts:
		m, err = f.NewCounterVec(opts, labelNames...)

	case *prometheus.CounterOpts:
		m, err = f.NewCounterVec(*opts, labelNames...)

	case prometheus.GaugeOpts:
		m, err = f.NewGaugeVec(opts, labelNames...)

	case *prometheus.GaugeOpts:
		m, err = f.NewGaugeVec(*opts, labelNames...)

	case prometheus.HistogramOpts:
		m, err = f.NewHistogramVec(opts, labelNames...)

	case *prometheus.HistogramOpts:
		m, err = f.NewHistogramVec(*opts, labelNames...)

	case prometheus.SummaryOpts:
		m, err = f.NewSummaryVec(opts, labelNames...)

	case *prometheus.SummaryOpts:
		m, err = f.NewSummaryVec(*opts, labelNames...)

	default:
		panic(fmt.Errorf("%T is not a recognized prometheus xxxOpts struct", o))
	}

	return
}

// NewCounter creates and registers a new counter using the supplied options.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewCounter
func (f *Factory) NewCounter(o prometheus.CounterOpts) (m prometheus.Counter, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewCounter(o)
		err = f.registerer.Register(m)
	}

	return
}

// NewCounterFunc creates and registers a new counter backed by the given function.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewCounterFunc
func (f *Factory) NewCounterFunc(o prometheus.CounterOpts, fn func() float64) (m prometheus.CounterFunc, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewCounterFunc(o, fn)
		err = f.registerer.Register(m)
	}

	return
}

// NewCounterVec creates and registers a new counter vector using the supplied options.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewCounterVec
func (f *Factory) NewCounterVec(o prometheus.CounterOpts, labelNames ...string) (m *prometheus.CounterVec, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewCounterVec(o, labelNames)
		err = f.registerer.Register(m)
	}

	return
}

// NewGauge creates and registers a new gauge using the supplied options.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewGauge
func (f *Factory) NewGauge(o prometheus.GaugeOpts) (m prometheus.Gauge, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewGauge(o)
		err = f.registerer.Register(m)
	}

	return
}

// NewGaugeFunc creates and registers a new gauge backed by the given function.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewGaugeFunc
func (f *Factory) NewGaugeFunc(o prometheus.GaugeOpts, fn func() float64) (m prometheus.GaugeFunc, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewGaugeFunc(o, fn)
		err = f.registerer.Register(m)
	}

	return
}

// NewGaugeVec creates and registers a new gauge vector using the supplied options.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewGaugeVec
func (f *Factory) NewGaugeVec(o prometheus.GaugeOpts, labelNames ...string) (m *prometheus.GaugeVec, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewGaugeVec(o, labelNames)
		err = f.registerer.Register(m)
	}

	return
}

// NewUntypedFunc creates and registers a new metric backed by the given function.
// The function must be of a signature supported by the package-level NewUntypedFunc.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewUntypedFunc
func (f *Factory) NewUntypedFunc(o prometheus.UntypedOpts, fn interface{}) (m prometheus.UntypedFunc, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)
		m, err = NewUntypedFunc(o, fn)
	}

	if err == nil {
		err = f.registerer.Register(m)
	}

	return
}

// NewHistogram creates and registers a new observer using the supplied options.
// The Observer component is backed by a prometheus.Histogram.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewHistogram
func (f *Factory) NewHistogram(o prometheus.HistogramOpts) (m prometheus.Observer, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		h := prometheus.NewHistogram(o)
		m = h
		err = f.registerer.Register(h)
	}

	return
}

// NewHistogramVec creates and registers a new observer vector using the supplied options.
// The ObserverVec component is backed by a prometheus.HistogramVec.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewHistogramVec
func (f *Factory) NewHistogramVec(o prometheus.HistogramOpts, labelNames ...string) (m prometheus.ObserverVec, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		h := prometheus.NewHistogramVec(o, labelNames)
		m = h
		err = f.registerer.Register(h)
	}

	return
}

// NewSummary creates and registers a new observer using the supplied options.
// The Observer component is backed by a prometheus.Summary.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewSummary
func (f *Factory) NewSummary(o prometheus.SummaryOpts) (m prometheus.Observer, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		s := prometheus.NewSummary(o)
		m = s
		err = f.registerer.Register(s)
	}

	return
}

// NewSummaryVec creates and registers a new observer vector using the supplied options.
// The ObserverVec component is backed by a prometheus.SummaryVec.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewSummaryVec
func (f *Factory) NewSummaryVec(o prometheus.SummaryOpts, labelNames ...string) (m prometheus.ObserverVec, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		ApplyDefaults(&o, f.defaults)
		f.warnOnNoHelp(o.Name, o.Help)

		s := prometheus.NewSummaryVec(o, labelNames)
		m = s
		err = f.registerer.Register(s)
	}

	return
}

// NewObserver creates a histogram or a summary, depending on the concrete type
// of the first parameter.  This method panics if o is not a prometheus.HistogramOpts
// or a prometheus.SummaryOpts.
func (f *Factory) NewObserver(o interface{}) (m prometheus.Observer, err error) {
	switch opts := o.(type) {
	case prometheus.HistogramOpts:
		m, err = f.NewHistogram(opts)

	case *prometheus.HistogramOpts:
		m, err = f.NewHistogram(*opts)

	case prometheus.SummaryOpts:
		m, err = f.NewSummary(opts)

	case *prometheus.SummaryOpts:
		m, err = f.NewSummary(*opts)

	default:
		panic(fmt.Errorf("%T is not a prometheus.HistogramOpts, a prometheus.SummaryOpts, or a pointer to either", o))
	}

	return
}

// NewObserverVec creates a histogram vector or a summary vector, depending on the concrete
// type of the first parameter.  This method panics if o is not a prometheus.HistogramOpts,
// a prometheus.SummaryOpts, or a pointer to either.
func (f *Factory) NewObserverVec(o interface{}, labelNames ...string) (m prometheus.ObserverVec, err error) {
	switch opts := o.(type) {
	case prometheus.HistogramOpts:
		m, err = f.NewHistogramVec(opts, labelNames...)

	case *prometheus.HistogramOpts:
		m, err = f.NewHistogramVec(*opts, labelNames...)

	case prometheus.SummaryOpts:
		m, err = f.NewSummaryVec(opts, labelNames...)

	case *prometheus.SummaryOpts:
		m, err = f.NewSummaryVec(*opts, labelNames...)

	default:
		panic(fmt.Errorf("%T is not a prometheus.HistogramOpts, a prometheus.SummaryOpts, or a pointer to either", o))
	}

	return
}
