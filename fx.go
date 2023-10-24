// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// Module is the name of the fx module that touchstone components
	// are provided within.
	Module = "touchstone"
)

// In represents the components used by this package to bootstrap
// a prometheus environment.  Provide uses these components.
type In struct {
	fx.In

	// Config is the prometheus configuration.  This is optional,
	// as a zero value for Config will result in a default environment.
	Config Config `optional:"true"`

	// Logger is the *zap.Logger to which this package writes messages.
	// This is optional, and if unset no messages are written.
	Logger *zap.Logger `optional:"true"`
}

// Provide bootstraps a prometheus environment for an uber/fx App.
// The following component types are provided by this function:
//
//   - prometheus.Gatherer
//   - promtheus.Registerer
//     NOTE: Do not rely on the Registerer actually being a *prometheus.Registry.
//     It may be decorated to arbitrary depth.
//   - *touchstone.Factory
func Provide() fx.Option {
	return fx.Module(
		Module,
		fx.Provide(
			func(in In) (prometheus.Gatherer, prometheus.Registerer, error) {
				return New(in.Config)
			},
			func(r prometheus.Registerer, in In) *Factory {
				return NewFactory(in.Config, in.Logger, r)
			},
		),
	)
}

// Metric emits a named component using the specified target.  The target
// is expected to be a function (constructor) of the same form accepted
// by fx.Annotated.Target.
//
// If name is empty, application startup is short-circuited with an error.
//
// See: https://pkg.go.dev/go.uber.org/fx#Annotated
func Metric(name string, target interface{}) fx.Option {
	if len(name) == 0 {
		return fx.Error(ErrNoMetricName)
	}

	return fx.Provide(
		fx.Annotated{
			Name:   name,
			Target: target,
		},
	)
}

// Counter uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.Counter with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func Counter(o prometheus.CounterOpts) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.Counter, error) {
			return f.NewCounter(o)
		},
	)
}

// CounterVec uses a Factory instance from the enclosing fx.App to create and register
// a *prometheus.CounterVec with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func CounterVec(o prometheus.CounterOpts, labelNames ...string) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (*prometheus.CounterVec, error) {
			return f.NewCounterVec(o, labelNames...)
		},
	)
}

// Gauge uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.Gauge with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func Gauge(o prometheus.GaugeOpts) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.Gauge, error) {
			return f.NewGauge(o)
		},
	)
}

// GaugeVec uses a Factory instance from the enclosing fx.App to create and register
// a *prometheus.GaugeVec with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func GaugeVec(o prometheus.GaugeOpts, labelNames ...string) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (*prometheus.GaugeVec, error) {
			return f.NewGaugeVec(o, labelNames...)
		},
	)
}

// Histogram uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.Observer with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func Histogram(o prometheus.HistogramOpts) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.Observer, error) {
			return f.NewHistogram(o)
		},
	)
}

// HistogramVec uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.ObserverVec with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func HistogramVec(o prometheus.HistogramOpts, labelNames ...string) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.ObserverVec, error) {
			return f.NewHistogramVec(o, labelNames...)
		},
	)
}

// Summary uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.Observer with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func Summary(o prometheus.SummaryOpts) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.Observer, error) {
			return f.NewSummary(o)
		},
	)
}

// SummaryVec uses a Factory instance from the enclosing fx.App to create and register
// a prometheus.ObserverVec with the same component name as the metric Name.
//
// If no Name is set, application startup is short-circuited with an error.
func SummaryVec(o prometheus.SummaryOpts, labelNames ...string) fx.Option {
	return Metric(
		o.Name,
		func(f *Factory) (prometheus.ObserverVec, error) {
			return f.NewSummaryVec(o, labelNames...)
		},
	)
}
