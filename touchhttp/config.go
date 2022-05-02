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
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

const (
	// HTTPErrorOnError is the value allowed for Config.ErrorHandling that maps
	// to promhttp.HTTPErrorOnError.  This is also the default used when no
	// value is set.
	HTTPErrorOnError = "http"

	// ContinueOnError is the value allowed for Config.ErrorHandler that maps
	// to promhttp.ContinueOnError.
	ContinueOnError = "continue"

	// PanicOnError is the value allowed for Config.ErrorHandler that maps
	// to promhttp.PanicOnError.
	PanicOnError = "panic"
)

// ErrorPrinter adapts an fx.Printer and allows it to be used as
// an error Logger for prometheus.
type ErrorPrinter struct {
	fx.Printer
}

// Println satisfies the promhttp.Logger interface.
func (ep ErrorPrinter) Println(values ...interface{}) {
	var msg strings.Builder

	// we want Fprintln behavior, but we don't want the trailing newline
	for i, v := range values {
		if i > 0 {
			msg.WriteRune(' ')
		}

		fmt.Fprint(&msg, v)
	}

	ep.Printer.Printf(msg.String())
}

// InvalidErrorHandlingError is the error returned when Config.ErrorHandling
// is not a recognized value.
type InvalidErrorHandlingError struct {
	// Value is the unrecognized Config.ErrorHandling value.
	Value string
}

// Error satisfies the error interface.
func (e *InvalidErrorHandlingError) Error() string {
	return fmt.Sprintf("Invalid ErrorHandling value: %s", e.Value)
}

// Config is the configuration for boostrapping the promhttp package.
type Config struct {
	// ErrorHandling is the promhttp.HandlerErrorHandling value.  If this field
	// is unset, promhttp.HTTPErrorOnError is used.
	//
	// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#HandlerErrorHandling
	ErrorHandling string `json:"errorHandling" yaml:"errorHandling"`

	// DisableCompression disables compression on metrics output.
	DisableCompression bool `json:"disableCompression" yaml:"disableCompression"`

	// MaxRequestsInFlight controls the number of concurrent HTTP metrics requests.
	MaxRequestsInFlight int `json:"maxRequestsInFlight" yaml:"maxRequestsInFlight"`

	// Timeout is the time period after which the handler will return a 503.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// EnableOpenMetrics controls whether open metrics encoding is available
	// during content negotiation.
	EnableOpenMetrics bool `json:"enableOpenMetrics" yaml:"enableOpenMetrics"`

	// InstrumentMetricHandler indicates whether the http.Handler that renders
	// prometheus metrics will itself be decorated with metrics.
	//
	// See: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp#InstrumentMetricHandler
	InstrumentMetricHandler bool `json:"instrumentMetricHandler" yaml:"instrumentMetricHandler"`
}

// NewHandlerOpts creates a basic HandlerOpts from an Config configuration.
func NewHandlerOpts(cfg Config, p fx.Printer, r prometheus.Registerer) (opts promhttp.HandlerOpts, err error) {
	opts = promhttp.HandlerOpts{
		DisableCompression:  cfg.DisableCompression,
		MaxRequestsInFlight: cfg.MaxRequestsInFlight,
		Timeout:             cfg.Timeout,
		EnableOpenMetrics:   cfg.EnableOpenMetrics,
		Registry:            r,
	}

	if p != nil {
		opts.ErrorLog = ErrorPrinter{Printer: p}
	}

	switch cfg.ErrorHandling {
	case "":
		// NOTE: Can just take the zero value here, since this package uses the same
		// default value as promhttp

	case HTTPErrorOnError:
		// this is just explicitly setting the error handling to the default

	case ContinueOnError:
		opts.ErrorHandling = promhttp.ContinueOnError

	case PanicOnError:
		opts.ErrorHandling = promhttp.PanicOnError

	default:
		err = &InvalidErrorHandlingError{Value: cfg.ErrorHandling}
	}

	return
}
