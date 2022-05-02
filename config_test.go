/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
