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
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type MetricTestSuite struct {
	suite.Suite
}

func (suite *MetricTestSuite) TestCounter() {
	app := fxtest.New(
		suite.T(),
		touchstone.Provide(),
		Counter(prometheus.CounterOpts{
			Name: "counter",
		}, "label"),
		fx.Invoke(
			func(in struct {
				fx.In
				Counter metrics.Counter `name:"counter"`
			}) {
				// verify the metric is usable
				in.Counter.With("label", "value").Add(1.0)
			},
		),
	)

	suite.Require().NoError(app.Err())
	app.RequireStart()
	app.RequireStop()
}

func (suite *MetricTestSuite) TestGauge() {
	app := fxtest.New(
		suite.T(),
		touchstone.Provide(),
		Gauge(prometheus.GaugeOpts{
			Name: "gauge",
		}, "label"),
		fx.Invoke(
			func(in struct {
				fx.In
				Gauge metrics.Gauge `name:"gauge"`
			}) {
				// verify the metric is usable
				in.Gauge.With("label", "value").Add(1.0)
			},
		),
	)

	suite.Require().NoError(app.Err())
	app.RequireStart()
	app.RequireStop()
}

func (suite *MetricTestSuite) TestHistogram() {
	app := fxtest.New(
		suite.T(),
		touchstone.Provide(),
		Histogram(prometheus.HistogramOpts{
			Name: "histogram",
		}, "label"),
		fx.Invoke(
			func(in struct {
				fx.In
				Histogram metrics.Histogram `name:"histogram"`
			}) {
				// verify the metric is usable
				in.Histogram.With("label", "value").Observe(10.5)
			},
		),
	)

	suite.Require().NoError(app.Err())
	app.RequireStart()
	app.RequireStop()
}

func (suite *MetricTestSuite) TestSummary() {
	app := fxtest.New(
		suite.T(),
		touchstone.Provide(),
		Summary(prometheus.SummaryOpts{
			Name: "summary",
		}, "label"),
		fx.Invoke(
			func(in struct {
				fx.In
				Summary metrics.Histogram `name:"summary"`
			}) {
				// verify the metric is usable
				in.Summary.With("label", "value").Observe(10.5)
			},
		),
	)

	suite.Require().NoError(app.Err())
	app.RequireStart()
	app.RequireStop()
}

func TestMetric(t *testing.T) {
	suite.Run(t, new(MetricTestSuite))
}
