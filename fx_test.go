package touchstone

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
)

type ProvideTestSuite struct {
	FxTestSuite
}

func (suite *ProvideTestSuite) TestDefaults() {
	var (
		gatherer   prometheus.Gatherer
		registerer prometheus.Registerer
		factory    *Factory
	)

	app := suite.newTestApp(
		Provide(),
		fx.Populate(
			&gatherer,
			&registerer,
			&factory,
		),
	)

	app.RequireStart()

	suite.NotNil(gatherer)
	suite.NotNil(registerer)
	suite.Require().NotNil(factory)
	suite.Empty(factory.DefaultNamespace())
	suite.Empty(factory.DefaultSubsystem())

	suite.Error(
		registerer.Register(prometheus.NewGoCollector()),
	)

	suite.Error(
		registerer.Register(prometheus.NewProcessCollector(
			prometheus.ProcessCollectorOpts{},
		)),
	)

	app.RequireStop()
}

func (suite *ProvideTestSuite) TestCustom() {
	var (
		gatherer   prometheus.Gatherer
		registerer prometheus.Registerer
		factory    *Factory
	)

	app := suite.newTestApp(
		fx.Supply(
			Config{
				DefaultNamespace:        "n",
				DefaultSubsystem:        "s",
				Pedantic:                true,
				DisableGoCollector:      true,
				DisableProcessCollector: true,
			},
		),
		Provide(),
		fx.Populate(
			&gatherer,
			&registerer,
			&factory,
		),
	)

	app.RequireStart()

	suite.NotNil(gatherer)
	suite.NotNil(registerer)
	suite.Require().NotNil(factory)
	suite.Equal("n", factory.DefaultNamespace())
	suite.Equal("s", factory.DefaultSubsystem())

	suite.NoError(
		registerer.Register(prometheus.NewGoCollector()),
	)

	suite.NoError(
		registerer.Register(prometheus.NewProcessCollector(
			prometheus.ProcessCollectorOpts{},
		)),
	)

	app.RequireStop()
}

func TestProvide(t *testing.T) {
	suite.Run(t, new(ProvideTestSuite))
}

type MetricTestSuite struct {
	FxTestSuite
}

// testMissingName verifies that a metric without a name short-circuits app startup
func (suite *MetricTestSuite) testMissingName(o fx.Option) {
	app := suite.newApp(
		Provide(),
		o,
	)

	suite.ErrorIs(app.Err(), ErrNoMetricName)
}

// testSuccess verifies that a metric got created.  the f parameter is an invoke
// function that is expected to do verification specific to the type of metric.
func (suite *MetricTestSuite) testSuccess(o fx.Option, f interface{}) {
	app := suite.newTestApp(
		Provide(),
		o,
		fx.Invoke(f),
	)

	app.RequireStart()
	app.RequireStop()
}

func (suite *MetricTestSuite) TestCounter() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			Counter(prometheus.CounterOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			Counter(prometheus.CounterOpts{
				Name: "test",
			}),
			func(in struct {
				fx.In
				Metric prometheus.Counter `name:"test"`
			}) {
				suite.Require().NotNil(in.Metric)
				in.Metric.Inc()
			},
		)
	})
}

func (suite *MetricTestSuite) TestCounterVec() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			CounterVec(prometheus.CounterOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			CounterVec(prometheus.CounterOpts{
				Name: "test",
			}, "label1"),
			func(in struct {
				fx.In
				Metric *prometheus.CounterVec `name:"test"`
			}) {
				m, err := in.Metric.GetMetricWith(prometheus.Labels{"label1": "value1"})
				suite.NoError(err)
				suite.Require().NotNil(m)
				m.Inc()
			},
		)
	})
}

func (suite *MetricTestSuite) TestGauge() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			Gauge(prometheus.GaugeOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			Gauge(prometheus.GaugeOpts{
				Name: "test",
			}),
			func(in struct {
				fx.In
				Metric prometheus.Gauge `name:"test"`
			}) {
				suite.Require().NotNil(in.Metric)
				in.Metric.Inc()
			},
		)
	})
}

func (suite *MetricTestSuite) TestGaugeVec() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			GaugeVec(prometheus.GaugeOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			GaugeVec(prometheus.GaugeOpts{
				Name: "test",
			}, "label1"),
			func(in struct {
				fx.In
				Metric *prometheus.GaugeVec `name:"test"`
			}) {
				m, err := in.Metric.GetMetricWith(prometheus.Labels{"label1": "value1"})
				suite.NoError(err)
				suite.Require().NotNil(m)
				m.Inc()
			},
		)
	})
}

func (suite *MetricTestSuite) TestHistogram() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			Histogram(prometheus.HistogramOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			Histogram(prometheus.HistogramOpts{
				Name: "test",
			}),
			func(in struct {
				fx.In
				Metric prometheus.Observer `name:"test"`
			}) {
				suite.Require().NotNil(in.Metric)
				in.Metric.Observe(1.0)
			},
		)
	})
}

func (suite *MetricTestSuite) TestHistogramVec() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			HistogramVec(prometheus.HistogramOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			HistogramVec(prometheus.HistogramOpts{
				Name: "test",
			}, "label1"),
			func(in struct {
				fx.In
				Metric prometheus.ObserverVec `name:"test"`
			}) {
				m, err := in.Metric.GetMetricWith(prometheus.Labels{"label1": "value1"})
				suite.NoError(err)
				suite.Require().NotNil(m)
				m.Observe(1.0)
			},
		)
	})
}

func (suite *MetricTestSuite) TestSummary() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			Summary(prometheus.SummaryOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			Summary(prometheus.SummaryOpts{
				Name: "test",
			}),
			func(in struct {
				fx.In
				Metric prometheus.Observer `name:"test"`
			}) {
				suite.Require().NotNil(in.Metric)
				in.Metric.Observe(1.0)
			},
		)
	})
}

func (suite *MetricTestSuite) TestSummaryVec() {
	suite.Run("MissingName", func() {
		suite.testMissingName(
			SummaryVec(prometheus.SummaryOpts{}),
		)
	})

	suite.Run("Success", func() {
		suite.testSuccess(
			SummaryVec(prometheus.SummaryOpts{
				Name: "test",
			}, "label1"),
			func(in struct {
				fx.In
				Metric prometheus.ObserverVec `name:"test"`
			}) {
				m, err := in.Metric.GetMetricWith(prometheus.Labels{"label1": "value1"})
				suite.NoError(err)
				suite.Require().NotNil(m)
				m.Observe(1.0)
			},
		)
	})
}

func TestMetric(t *testing.T) {
	suite.Run(t, new(MetricTestSuite))
}
