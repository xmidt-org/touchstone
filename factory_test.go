package touchstone

import (
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone/touchtest"
)

type FactoryTestSuite struct {
	suite.Suite
}

// Printf allows this suite to be used directly as an fx.Printer.
func (suite *FactoryTestSuite) Printf(format string, arguments ...interface{}) {
	suite.T().Logf(format, arguments...)
}

func (suite *FactoryTestSuite) newFactory(cfg Config) (*Factory, prometheus.Gatherer, prometheus.Registerer) {
	g, r, err := New(cfg)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	f := NewFactory(cfg, suite, r)
	suite.Require().NotNil(f)
	suite.Equal(cfg.DefaultNamespace, f.DefaultNamespace())
	suite.Equal(cfg.DefaultSubsystem, f.DefaultSubsystem())
	return f, g, r
}

func (suite *FactoryTestSuite) newAssertions(g prometheus.Gatherer) *touchtest.Assertions {
	a := touchtest.NewSuite(suite)
	return a.Expect(g)
}

// labelsPresent checks that the given metric has a series of label names.
// Note that this method typically must be used before a *touchtest.Assertions,
// since the act of getting a metric with labels causes it to show up in the gatherer.
func (suite *FactoryTestSuite) labelsPresent(v interface{}, labelNames ...string) {
	l := make(prometheus.Labels, len(labelNames))
	for i, labelName := range labelNames {
		l[labelName] = "value" + strconv.Itoa(i)
	}

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

func (suite *FactoryTestSuite) TestNew() {
	suite.Run("InvalidOpts", func() {
		f, _, _ := suite.newFactory(Config{})
		suite.Panics(func() {
			f.New(123)
		})
	})

	testCases := []struct {
		metricName string
		opts       interface{}
		implements interface{}
	}{
		{
			metricName: "test",
			opts: prometheus.CounterOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Counter)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.CounterOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Counter)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.GaugeOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Gauge)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.GaugeOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Gauge)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Histogram)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Histogram)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Summary)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Summary)(nil),
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			f, g, _ := suite.newFactory(Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			})

			m, err := f.New(testCase.opts)
			suite.NoError(err)
			suite.NotNil(m)
			suite.Implements(testCase.implements, m)

			ma := suite.newAssertions(g)
			ma.Registered(prometheus.BuildFQName("n", "s", testCase.metricName))
		})
	}
}

func (suite *FactoryTestSuite) TestNewVec() {
	suite.Run("InvalidOpts", func() {
		f, _, _ := suite.newFactory(Config{})
		suite.Panics(func() {
			f.NewVec(123)
		})
	})

	testCases := []struct {
		metricName string
		opts       interface{}
		isType     interface{}
	}{
		{
			metricName: "test",
			opts: prometheus.CounterOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.CounterVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.CounterOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.CounterVec)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.GaugeOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.GaugeVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.GaugeOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.GaugeVec)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.HistogramVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.HistogramVec)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.SummaryVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.SummaryVec)(nil),
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			f, g, _ := suite.newFactory(Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			})

			m, err := f.NewVec(testCase.opts, "label1", "label2")
			suite.NoError(err)
			suite.NotNil(m)
			suite.IsType(testCase.isType, m)
			suite.labelsPresent(m, "label1", "label2")

			ma := suite.newAssertions(g)
			ma.Registered(prometheus.BuildFQName("n", "s", testCase.metricName))
		})
	}
}

func (suite *FactoryTestSuite) TestNewCounter() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewCounter(prometheus.CounterOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewCounter(prometheus.CounterOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounter(prometheus.CounterOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounter(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewCounterFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewCounterFunc(prometheus.CounterOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewCounterFunc(prometheus.CounterOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterFunc(prometheus.CounterOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterFunc(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewCounterVec() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewCounterVec(prometheus.CounterOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewCounterVec(prometheus.CounterOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterVec(prometheus.CounterOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewCounterVec(
			prometheus.CounterOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewGauge() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewGauge(prometheus.GaugeOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewGauge(prometheus.GaugeOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGauge(prometheus.GaugeOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGauge(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewGaugeFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeFunc(prometheus.GaugeOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeFunc(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewGaugeVec() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeVec(prometheus.GaugeOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewGaugeVec(prometheus.GaugeOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeVec(prometheus.GaugeOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewGaugeVec(
			prometheus.GaugeOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewUntypedFunc() {
	fn := func() float64 { return 1.0 }

	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{}, fn)
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewUntypedFunc(prometheus.UntypedOpts{Name: "test"}, fn)
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewUntypedFunc(
			prometheus.UntypedOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			fn,
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewHistogram() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewHistogram(prometheus.HistogramOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewHistogram(prometheus.HistogramOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogram(prometheus.HistogramOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogram(
			prometheus.HistogramOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewHistogramVec() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewHistogramVec(prometheus.HistogramOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewHistogramVec(prometheus.HistogramOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogramVec(prometheus.HistogramOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewHistogramVec(
			prometheus.HistogramOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewSummary() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewSummary(prometheus.SummaryOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewSummary(prometheus.SummaryOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummary(prometheus.SummaryOpts{Name: "test"})
		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummary(
			prometheus.SummaryOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
		)

		suite.NoError(err)
		suite.NotNil(m)

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewSummaryVec() {
	suite.Run("NoName", func() {
		f, _, _ := suite.newFactory(Config{})
		m, err := f.NewSummaryVec(prometheus.SummaryOpts{})
		suite.ErrorIs(err, ErrNoMetricName)
		suite.Nil(m)
	})

	suite.Run("NoDefaults", func() {
		f, g, _ := suite.newFactory(Config{})
		m, err := f.NewSummaryVec(prometheus.SummaryOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered("test")
	})

	suite.Run("Defaults", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummaryVec(prometheus.SummaryOpts{Name: "test"}, "label1")
		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("n", "s", "test"))
		ma.NotRegistered("test")
	})

	suite.Run("Overrides", func() {
		f, g, _ := suite.newFactory(Config{
			DefaultNamespace: "n",
			DefaultSubsystem: "s",
		})

		m, err := f.NewSummaryVec(
			prometheus.SummaryOpts{Namespace: "o1", Subsystem: "o2", Name: "test", Help: "some lovely help"},
			"label1",
		)

		suite.NoError(err)
		suite.labelsPresent(m, "label1")

		ma := suite.newAssertions(g)
		ma.Registered(prometheus.BuildFQName("o1", "o2", "test"))
		ma.NotRegistered("test")
		ma.NotRegistered(prometheus.BuildFQName("n", "s", "test"))
	})
}

func (suite *FactoryTestSuite) TestNewObserver() {
	suite.Run("InvalidOpts", func() {
		f, _, _ := suite.newFactory(Config{})
		suite.Panics(func() {
			f.NewObserver(prometheus.CounterOpts{})
		})
	})

	testCases := []struct {
		metricName string
		opts       interface{}
		implements interface{}
	}{
		{
			metricName: "test",
			opts: prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Histogram)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Histogram)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Summary)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			implements: (*prometheus.Summary)(nil),
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			f, g, _ := suite.newFactory(Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			})

			m, err := f.NewObserver(testCase.opts)
			suite.NoError(err)
			suite.NotNil(m)
			suite.Implements(testCase.implements, m)

			ma := suite.newAssertions(g)
			ma.Registered(prometheus.BuildFQName("n", "s", testCase.metricName))
		})
	}
}

func (suite *FactoryTestSuite) TestNewObserverVec() {
	suite.Run("InvalidOpts", func() {
		f, _, _ := suite.newFactory(Config{})
		suite.Panics(func() {
			f.NewObserverVec(prometheus.CounterOpts{})
		})
	})

	testCases := []struct {
		metricName string
		opts       interface{}
		isType     interface{}
	}{
		{
			metricName: "test",
			opts: prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.HistogramVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.HistogramOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.HistogramVec)(nil),
		},
		{
			metricName: "test",
			opts: prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.SummaryVec)(nil),
		},
		{
			metricName: "test",
			opts: &prometheus.SummaryOpts{
				Name: "test",
				Help: "test",
			},
			isType: (*prometheus.SummaryVec)(nil),
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			f, g, _ := suite.newFactory(Config{
				DefaultNamespace: "n",
				DefaultSubsystem: "s",
			})

			m, err := f.NewObserverVec(testCase.opts, "label1", "label2")
			suite.NoError(err)
			suite.NotNil(m)
			suite.IsType(testCase.isType, m)
			suite.labelsPresent(m, "label1", "label2")

			ma := suite.newAssertions(g)
			ma.Registered(prometheus.BuildFQName("n", "s", testCase.metricName))
		})
	}
}

func TestFactory(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
