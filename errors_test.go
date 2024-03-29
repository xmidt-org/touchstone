// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

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
