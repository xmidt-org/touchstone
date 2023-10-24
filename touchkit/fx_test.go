// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
