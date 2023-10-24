// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

// Handler is a type alias for http.Handler that makes dependency injection easier.
// The handler bootstrapped by this package will be of this type, which means
// injection by type will not interfere with any other http.Handler component in
// the application.
//
// The promhttp.HandlerFor function is used to create this type using the handler
// options created from the Config type.
//
// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#HandlerFor
type Handler http.Handler

// In represents the components used by this package to bootstrap
// a promhttp environment.  Provide uses these components.
type In struct {
	fx.In

	// Config is the prometheus configuration.  This is optional,
	// as a zero value for Config will result in a default environment.
	Config Config `optional:"true"`

	// Printer is the fx.Printer to which this package writes messages.
	// This is optional, and if unset no messages are written.
	Printer fx.Printer `optional:"true"`

	// Now is the optional current time function.  If supplied, this
	// will be used for computing metric durations.
	Now func() time.Time `optional:"true"`
}

// Provide bootstraps the promhttp environment for an uber/fx app.  This
// function creates the following component types:
//
//   - promhttp.HandlerOpts
//   - touchhttp.Handler
//     This is the http.Handler to use to serve prometheus metrics.
//     It will be instrumented if Config.InstrumentMetricHandler is set to true.
func Provide() fx.Option {
	return fx.Provide(
		func(r prometheus.Registerer, in In) (promhttp.HandlerOpts, error) {
			return NewHandlerOpts(in.Config, in.Printer, r)
		},
		func(r prometheus.Registerer, g prometheus.Gatherer, opts promhttp.HandlerOpts, in In) (h Handler) {
			h = promhttp.HandlerFor(g, opts)
			if in.Config.InstrumentMetricHandler {
				h = promhttp.InstrumentMetricHandler(r, h)
			}

			return
		},
	)
}
