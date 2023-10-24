// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/suite"
)

type NewTestSuite struct {
	suite.Suite
}

func (suite *NewTestSuite) TestDefault() {
	g, r, err := New(Config{})
	suite.NoError(err)
	suite.NotNil(g)
	suite.NotNil(r)

	suite.Error(
		// the go collector should have been registered
		r.Register(collectors.NewGoCollector()),
	)

	suite.Error(
		// the process collector should have been registered
		r.Register(collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		)),
	)
}

func (suite *NewTestSuite) TestPedantic() {
	g, r, err := New(Config{Pedantic: true})
	suite.NoError(err)
	suite.NotNil(g)
	suite.NotNil(r)

	suite.Error(
		// the go collector should have been registered
		r.Register(collectors.NewGoCollector()),
	)

	suite.Error(
		// the process collector should have been registered
		r.Register(collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		)),
	)
}

func (suite *NewTestSuite) TestDisableStandardCollectors() {
	g, r, err := New(Config{
		DisableGoCollector:        true,
		DisableProcessCollector:   true,
		DisableBuildInfoCollector: true,
	})

	suite.NoError(err)
	suite.NotNil(g)
	suite.NotNil(r)

	suite.NoError(
		r.Register(collectors.NewGoCollector()),
	)

	suite.NoError(
		r.Register(collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		)),
	)
}

func TestNew(t *testing.T) {
	suite.Run(t, new(NewTestSuite))
}
