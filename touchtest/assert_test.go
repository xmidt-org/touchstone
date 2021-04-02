package touchtest

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type AssertionsTestSuite struct {
	suite.Suite
}

func (suite *AssertionsTestSuite) register(r prometheus.Registerer, c ...prometheus.Collector) {
	for _, m := range c {
		suite.Require().NoError(r.Register(m))
	}
}

func (suite *AssertionsTestSuite) TestNewSuite() {
	a := NewSuite(suite)
	suite.NotNil(a.assert)
	suite.NotNil(a.require)
}

func (suite *AssertionsTestSuite) TestRegistered() {
	var (
		r  = prometheus.NewPedanticRegistry()
		mt = &mockTestingT{t: suite.T()}
		a  = New(mt)
	)

	suite.register(
		r,
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "counter",
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gauge",
		}),
	)

	suite.Same(a, a.Expect(r))

	suite.True(a.Registered("counter", "gauge"))
	suite.Zero(mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0

	suite.False(a.Registered("grumpkins", "snarks"))
	suite.Equal(2, mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0
}

func (suite *AssertionsTestSuite) TestNotRegistered() {
	var (
		r  = prometheus.NewPedanticRegistry()
		mt = &mockTestingT{t: suite.T()}
		a  = New(mt)
	)

	suite.register(
		r,
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "counter",
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gauge",
		}),
	)

	suite.Same(a, a.Expect(r))

	suite.False(a.NotRegistered("counter", "gauge"))
	suite.Equal(2, mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0

	suite.True(a.NotRegistered("grumpkins", "snarks"))
	suite.Zero(mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0
}

func (suite *AssertionsTestSuite) TestGatherAndCompare() {
	var (
		expected        = prometheus.NewPedanticRegistry()
		expectedCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "testCounter",
			Help: "testCounter",
		})
		expectedGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "testGauge",
			Help: "testGauge",
		})

		actual        = prometheus.NewPedanticRegistry()
		actualCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "testCounter",
			Help: "testCounter",
		})
		actualGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "testGauge",
			Help: "testGauge",
		})

		mt = &mockTestingT{t: suite.T()}
		a  = New(mt)
	)

	suite.register(expected, expectedCounter, expectedGauge)
	expectedCounter.Inc()
	expectedGauge.Add(2.0)

	a.Expect(expected)

	suite.register(actual, actualCounter, actualGauge)
	suite.False(a.GatherAndCompare(actual))
	suite.Equal(1, mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0

	actualCounter.Inc()
	actualGauge.Add(2.0)
	suite.True(a.GatherAndCompare(actual))
	suite.Zero(mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0
}

func (suite *AssertionsTestSuite) TestCollectAndCompare() {
	var (
		expected        = prometheus.NewPedanticRegistry()
		expectedCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "testCounter",
			Help: "testCounter",
		})

		actualCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "testCounter",
			Help: "testCounter",
		})

		mt = &mockTestingT{t: suite.T()}
		a  = New(mt)
	)

	suite.register(expected, expectedCounter)
	expectedCounter.Inc()

	a.Expect(expected)

	suite.False(a.CollectAndCompare(actualCounter))
	suite.Equal(1, mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0

	actualCounter.Inc()
	suite.True(a.CollectAndCompare(actualCounter))
	suite.Zero(mt.errors)
	suite.Zero(mt.failures)
	mt.errors = 0
	mt.failures = 0
}

func TestAssertions(t *testing.T) {
	suite.Run(t, new(AssertionsTestSuite))
}
