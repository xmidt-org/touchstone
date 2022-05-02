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

package touchkit

import (
	"github.com/go-kit/kit/metrics"
	promkit "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
)

// Counter uses an injected touchstone Factory to create a go-kit metrics.Counter backed
// by a prometheus CounterVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Counter(o prometheus.CounterOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *touchstone.Factory) (m metrics.Counter, err error) {
			var pm *prometheus.CounterVec
			pm, err = f.NewCounterVec(o, labelNames...)
			if err == nil {
				m = promkit.NewCounter(pm)
			}

			return
		},
	})
}

// Gauge uses an injected touchstone Factory to create a go-kit metrics.Gauge backed
// by a prometheus GaugeVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Gauge(o prometheus.GaugeOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *touchstone.Factory) (m metrics.Gauge, err error) {
			var pm *prometheus.GaugeVec
			pm, err = f.NewGaugeVec(o, labelNames...)
			if err == nil {
				m = promkit.NewGauge(pm)
			}

			return
		},
	})
}

// Histogram uses an injected touchstone Factory to create a go-kit metrics.Histogram backed
// by a prometheus HistogramVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Histogram(o prometheus.HistogramOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *touchstone.Factory) (m metrics.Histogram, err error) {
			var pm prometheus.ObserverVec
			pm, err = f.NewHistogramVec(o, labelNames...)
			if err == nil {
				m = promkit.NewHistogram(pm.(*prometheus.HistogramVec))
			}

			return
		},
	})
}

// Summary uses an injected touchstone Factory to create a go-kit metrics.Histogram backed
// by a prometheus SummaryVec.  The *touchstone.Factory from the enclosing fx.App
// is used to create and register the prometheus metric.  The name of the returned
// component will be the same as the metric name.
func Summary(o prometheus.SummaryOpts, labelNames ...string) fx.Option {
	return fx.Provide(fx.Annotated{
		Name: o.Name,
		Target: func(f *touchstone.Factory) (m metrics.Histogram, err error) {
			var pm prometheus.ObserverVec
			pm, err = f.NewSummaryVec(o, labelNames...)
			if err == nil {
				m = promkit.NewSummary(pm.(*prometheus.SummaryVec))
			}

			return
		},
	})
}
