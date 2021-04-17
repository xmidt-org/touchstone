package touchstone

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type CollectorAsTestSuite struct {
	suite.Suite
}

func (suite *CollectorAsTestSuite) TestNilInterface() {
	suite.Panics(func() {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test",
		})

		CollectorAs(c, nil)
	})
}

func (suite *CollectorAsTestSuite) TestNonPointer() {
	suite.Panics(func() {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test",
		})

		CollectorAs(c, 123)
	})
}

func (suite *CollectorAsTestSuite) TestNilPointer() {
	suite.Panics(func() {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test",
		})

		var thisIsNil *prometheus.Counter

		CollectorAs(c, thisIsNil)
	})
}

func (suite *CollectorAsTestSuite) TestInvalidTarget() {
	suite.Panics(func() {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test",
		})

		var invalid int
		CollectorAs(c, &invalid)
	})
}

func (suite *CollectorAsTestSuite) TestMetric() {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test",
	})

	var target prometheus.Counter
	suite.Require().True(CollectorAs(c, &target))
	suite.Equal(c, target)
}

func (suite *CollectorAsTestSuite) TestVector() {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test",
	}, []string{"label1"})

	var target *prometheus.CounterVec
	suite.Require().True(CollectorAs(cv, &target))
	suite.Equal(cv, target)
}

func (suite *CollectorAsTestSuite) TestCustomInterface() {
	type incrementer interface {
		Inc()
	}

	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test",
	})

	var target incrementer
	suite.Require().True(CollectorAs(c, &target))
	suite.Equal(c, target)
}

func (suite *CollectorAsTestSuite) TestWrongTargetType() {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test",
	}, []string{"label1"})

	var target *prometheus.GaugeVec
	suite.False(CollectorAs(cv, &target))
	suite.Nil(target)
}

func TestCollectorAs(t *testing.T) {
	suite.Run(t, new(CollectorAsTestSuite))
}

type AsAlreadyRegisteredErrorTestSuite struct {
	suite.Suite
}

func (suite *AsAlreadyRegisteredErrorTestSuite) TestNil() {
	suite.Nil(AsAlreadyRegisteredError(nil))
}

func (suite *AsAlreadyRegisteredErrorTestSuite) TestAlreadyRegisteredError() {
	input := prometheus.AlreadyRegisteredError{
		ExistingCollector: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "already_registered",
			Help: "existing",
		}),
		NewCollector: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "already_registered",
			Help: "new",
		}),
	}

	result := AsAlreadyRegisteredError(input)
	suite.Require().NotNil(result)
	suite.Equal(input, *result)
}

func (suite *AsAlreadyRegisteredErrorTestSuite) TestNotAlreadyRegisteredError() {
	input := errors.New("this is not an AlreadyRegisteredError")
	suite.Nil(AsAlreadyRegisteredError(input))
}

func TestAsAlreadyRegisteredError(t *testing.T) {
	suite.Run(t, new(AsAlreadyRegisteredErrorTestSuite))
}

type ExistingCollectorTestSuite struct {
	suite.Suite
}

func (suite *ExistingCollectorTestSuite) TestNilError() {
	suite.Run("NilTarget", func() {
		suite.Nil(ExistingCollector(nil, nil))
	})

	suite.Run("WithTarget", func() {
		var cv *prometheus.CounterVec
		suite.Nil(ExistingCollector(&cv, nil))
		suite.Nil(cv)
	})
}

func (suite *ExistingCollectorTestSuite) TestNotAlreadyRegisteredError() {
	suite.Run("NilTarget", func() {
		input := errors.New("this is not an AlreadyRegisteredError")
		suite.Equal(input, ExistingCollector(nil, input))
	})

	suite.Run("WithTarget", func() {
		input := errors.New("this is not an AlreadyRegisteredError")
		var cv *prometheus.CounterVec
		suite.Equal(input, ExistingCollector(nil, input))
		suite.Nil(cv)
	})
}

func (suite *ExistingCollectorTestSuite) TestAlreadyRegisteredError() {
	suite.Run("NilTarget", func() {
		input := prometheus.AlreadyRegisteredError{
			ExistingCollector: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "already_registered",
				Help: "existing",
			}),
			NewCollector: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "already_registered",
				Help: "new",
			}),
		}

		suite.Panics(func() {
			ExistingCollector(nil, input)
		})
	})

	suite.Run("WithTarget", func() {
		input := prometheus.AlreadyRegisteredError{
			ExistingCollector: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "already_registered",
				Help: "existing",
			}),
			NewCollector: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "already_registered",
				Help: "new",
			}),
		}

		var c prometheus.Counter
		suite.Nil(ExistingCollector(&c, input))
		suite.Equal(input.ExistingCollector, c)
	})
}

func TestExistingCollector(t *testing.T) {
	suite.Run(t, new(ExistingCollectorTestSuite))
}
