package touchhttp

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type ProvideTestSuite struct {
	suite.Suite
}

// assertMetricExists verifies that a metric exists by simply attempting to
// register a counter with the given name.
func (suite *ProvideTestSuite) assertMetricExists(r prometheus.Registerer, name, msg string) bool {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{Name: name},
	)

	return suite.Error(r.Register(c), msg)
}

// assertMetricNotExists verifies that a metric exists by simply attempting to
// register a counter with the given name.
func (suite *ProvideTestSuite) assertMetricNotExists(r prometheus.Registerer, name, msg string) bool {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{Name: name},
	)

	return suite.NoError(r.Register(c), msg)
}

func (suite *ProvideTestSuite) TestDefaults() {
	var (
		ho promhttp.HandlerOpts
		h  Handler
		sb ServerBundle
		cb ClientBundle
		r  prometheus.Registerer

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			Provide(),
			fx.Populate(&ho, &h, &sb, &cb, &r),
		)
	)

	suite.NoError(app.Err())
	app.RequireStart()

	suite.assertMetricExists(r, MetricServerRequestCount, "ServerBundle not initialized properly")
	suite.assertMetricExists(r, MetricClientRequestCount, "ClientBundle not initialized properly")

	suite.assertMetricNotExists(
		r,
		"promhttp_metric_handler_requests_total",
		"Handler should not have been instrumented",
	)

	app.RequireStop()
}

func (suite *ProvideTestSuite) TestInstrumentMetricHandler() {
	var (
		ho promhttp.HandlerOpts
		h  Handler
		sb ServerBundle
		cb ClientBundle
		r  prometheus.Registerer

		app = fxtest.New(
			suite.T(),
			fx.Supply(
				Config{
					InstrumentMetricHandler: true,
				},
			),
			touchstone.Provide(),
			Provide(),
			fx.Populate(&ho, &h, &sb, &cb, &r),
		)
	)

	suite.NoError(app.Err())
	app.RequireStart()

	suite.assertMetricExists(r, MetricServerRequestCount, "ServerBundle not initialized properly")
	suite.assertMetricExists(r, MetricClientRequestCount, "ClientBundle not initialized properly")

	suite.assertMetricExists(
		r,
		"promhttp_metric_handler_requests_total",
		"Handler should have been instrumented",
	)

	app.RequireStop()
}

func TestProvide(t *testing.T) {
	suite.Run(t, new(ProvideTestSuite))
}
