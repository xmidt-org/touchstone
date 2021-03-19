package touchstone

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type FactoryTestSuite struct {
	suite.Suite
}

// Printf allows this suite to be used directly as an fx.Printer.
func (suite *FactoryTestSuite) Printf(format string, arguments ...interface{}) {
	suite.T().Logf(format, arguments...)
}

func (suite *FactoryTestSuite) newFactory(cfg Config) (*Factory, prometheus.Registerer) {
	_, r, err := New(cfg)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	f := NewFactory(cfg, suite, r)
	suite.Require().NotNil(f)
	suite.Equal(cfg.DefaultNamespace, f.DefaultNamespace())
	suite.Equal(cfg.DefaultSubsystem, f.DefaultSubsystem())
	return f, r
}

func (suite *FactoryTestSuite) metricPresent(r prometheus.Registerer, o prometheus.Opts) {
	suite.Error(
		r.Register(
			// it doesn't matter what kind of metric we try to register,
			// as long as it's a duplicate
			prometheus.NewCounter(prometheus.CounterOpts(o)),
		),
		"A metric with the name [%s] SHOULD exist",
		prometheus.BuildFQName(o.Namespace, o.Subsystem, o.Name),
	)
}

func (suite *FactoryTestSuite) metricAbsent(r prometheus.Registerer, o prometheus.Opts) {
	suite.NoError(
		r.Register(
			// it doesn't matter what kind of metric we try to register,
			// as long as it's a duplicate
			prometheus.NewCounter(prometheus.CounterOpts(o)),
		),
		"A metric with the name [%s] SHOULD NOT exist",
		prometheus.BuildFQName(o.Namespace, o.Subsystem, o.Name),
	)
}

func (suite *FactoryTestSuite) labelsPresent(v interface{}, l prometheus.Labels) {
	suite.Require().NotNil(v)
	var vec *prometheus.MetricVec

	switch vt := v.(type) {
	case *prometheus.CounterVec:
		vec = vt.MetricVec

	case *prometheus.GaugeVec:
		vec = vt.MetricVec

	case prometheus.ObserverVec:
		m, err := vt.GetMetricWith(l)
		suite.NoError(err, "The metric should have had the labels: %v", l)
		suite.NotNil(m, "The metric should have had the labels: %v", l)
		return

	default:
		suite.Require().Fail("Invalid labelled metric type", "%T", v)
	}

	suite.Require().NotNil(vec)
	m, err := vec.GetMetricWith(l)
	suite.NoError(err, "The metric should have had the labels: %v", l)
	suite.NotNil(m, "The metric should have had the labels: %v", l)
}

func (suite *FactoryTestSuite) TestNewCounter() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewCounter(prometheus.CounterOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewCounter(prometheus.CounterOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounter(prometheus.CounterOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounter(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewCounterFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewCounterFunc(prometheus.CounterOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewCounterFunc(prometheus.CounterOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterFunc(prometheus.CounterOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterFunc(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewCounterVec() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewCounterVec(prometheus.CounterOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewCounterVec(prometheus.CounterOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterVec(prometheus.CounterOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterVec(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewGauge() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewGauge(prometheus.GaugeOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewGauge(prometheus.GaugeOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGauge(prometheus.GaugeOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGauge(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewGaugeFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeFunc(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewGaugeVec() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeVec(prometheus.GaugeOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewGaugeVec(prometheus.GaugeOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeVec(prometheus.GaugeOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeVec(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewUntypedFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewUntypedFunc(
			prometheus.UntypedOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewHistogram() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewHistogram(prometheus.HistogramOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewHistogram(prometheus.HistogramOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogram(prometheus.HistogramOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogram(
			prometheus.HistogramOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewHistogramVec() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewHistogramVec(prometheus.HistogramOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewHistogramVec(prometheus.HistogramOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogramVec(prometheus.HistogramOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogramVec(
			prometheus.HistogramOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewSummary() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewSummary(prometheus.SummaryOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewSummary(prometheus.SummaryOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummary(prometheus.SummaryOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummary(
			prometheus.SummaryOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func (suite *FactoryTestSuite) TestNewSummaryVec() {
	suite.Run("NoName", func() {
		f, _ := suite.newFactory(Config{})
		m, err := f.NewSummaryVec(prometheus.SummaryOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, r := suite.newFactory(Config{})
		m, err := f.NewSummaryVec(prometheus.SummaryOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "namespace", Name: "test"})
	})

	suite.Run("Defaults", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummaryVec(prometheus.SummaryOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
	})

	suite.Run("Overrides", func() {
		f, r := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummaryVec(
			prometheus.SummaryOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, prometheus.Labels{"label1": "value1"})

		suite.metricPresent(r, prometheus.Opts{Namespace: "o1", Subsystem: "o2", Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Name: "test"})
		suite.metricAbsent(r, prometheus.Opts{Namespace: "n", Subsystem: "s", Name: "test"})
	})
}

func TestFactory(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}