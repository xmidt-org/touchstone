package touchkit

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// Provide creates the *Factory that permits working directly with go-kit metrics
// in a manner similar to the touchstone package.
//
// This function requires the parent package to be bootstrapped.
func Provide() fx.Option {
	return fx.Provide(
		NewFactory,
	)
}

// Counter uses the injected Factory to create a go-kit metrics.Counter backed
// by a prometheus CounterVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Counter(o prometheus.CounterOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *Factory) (metrics.Counter, error) {
			return f.NewCounter(o, labelNames...)
		},
	})
}

// Gauge uses the injected Factory to create a go-kit metrics.Gauge backed
// by a prometheus GaugeVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Gauge(o prometheus.GaugeOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *Factory) (metrics.Gauge, error) {
			return f.NewGauge(o, labelNames...)
		},
	})
}

// Histogram uses the injected Factory to create a go-kit metrics.Histogram backed
// by a prometheus HistogramVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Histogram(o prometheus.HistogramOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *Factory) (metrics.Histogram, error) {
			return f.NewHistogram(o, labelNames...)
		},
	})
}

// Summary uses the injected Factory to create a go-kit metrics.Histogram backed
// by a prometheus SummaryVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Summary(o prometheus.SummaryOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *Factory) (metrics.Histogram, error) {
			return f.NewSummary(o, labelNames...)
		},
	})
}
