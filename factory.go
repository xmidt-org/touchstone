package touchstone

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
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
// If an fx.Printer is supplied, it is used to log warnings about missing Help
// in *Opts structs.
//
// This package's functions that match metric types, e.g. Counter, CounterVec, etc, use
// a Factory instance injected from the enclosing fx.App.  Those functions are generally
// preferred to using a Factory directly, since they emit their metrics as components which
// can then be injected as needed.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promauto
type Factory struct {
	defaultNamespace string
	defaultSubsystem string
	printer          fx.Printer
	registerer       prometheus.Registerer
}

// NewFactory produces a Factory that uses the supplied registry.
func NewFactory(cfg Config, p fx.Printer, r prometheus.Registerer) *Factory {
	return &Factory{
		defaultNamespace: cfg.DefaultNamespace,
		defaultSubsystem: cfg.DefaultSubsystem,
		printer:          p,
		registerer:       r,
	}
}

func (f *Factory) checkName(v string) error {
	if len(v) == 0 {
		return ErrNoMetricName
	}

	return nil
}

func (f *Factory) namespace(v string) string {
	if len(v) > 0 {
		return v
	}

	return f.defaultNamespace
}

func (f *Factory) subsystem(v string) string {
	if len(v) > 0 {
		return v
	}

	return f.defaultSubsystem
}

func (f *Factory) printf(format string, args ...interface{}) {
	if f.printer != nil {
		f.printer.Printf("[TOUCHSTONE] "+format, args...)
	}
}

func (f *Factory) warnOnNoHelp(name, help string) {
	if len(help) == 0 {
		f.printf("WARNING: No Help set for metric: %s", name)
	}
}

// DefaultNamespace returns the namespace used to register metrics
// when no Namespace is specified in the *Opts struct.  This may be
// empty to indicate that there is no default.
func (f *Factory) DefaultNamespace() string {
	return f.defaultNamespace
}

// DefaultSubsystem returns the subsystem used to register metrics
// when no Subsystem is specified in the *Opts struct.  This may be
// empty to indicate that there is no default.
func (f *Factory) DefaultSubsystem() string {
	return f.defaultSubsystem
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewGaugeVec(o, labelNames)
		err = f.registerer.Register(m)
	}

	return
}

// NewUntypedFunc creates and registers a new metric backed by the given function.
//
// This method returns an error if the options do not specify a name.  Both namespace
// and subsystem are defaulted appropriately if not set in the options.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#NewUntypedFunc
func (f *Factory) NewUntypedFunc(o prometheus.UntypedOpts, fn func() float64) (m prometheus.UntypedFunc, err error) {
	err = f.checkName(o.Name)
	if err == nil {
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
		f.warnOnNoHelp(o.Name, o.Help)

		m = prometheus.NewUntypedFunc(o, fn)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
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
		o.Namespace = f.namespace(o.Namespace)
		o.Subsystem = f.subsystem(o.Subsystem)
		f.warnOnNoHelp(o.Name, o.Help)

		s := prometheus.NewSummaryVec(o, labelNames)
		m = s
		err = f.registerer.Register(s)
	}

	return
}
