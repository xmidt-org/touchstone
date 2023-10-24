// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
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
