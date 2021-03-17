package touchstone

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
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
		r.Register(prometheus.NewGoCollector()),
	)

	suite.Error(
		// the process collector should have been registered
		r.Register(prometheus.NewProcessCollector(
			prometheus.ProcessCollectorOpts{},
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
		r.Register(prometheus.NewGoCollector()),
	)

	suite.Error(
		// the process collector should have been registered
		r.Register(prometheus.NewProcessCollector(
			prometheus.ProcessCollectorOpts{},
		)),
	)
}

func (suite *NewTestSuite) TestDisableStandardCollectors() {
	g, r, err := New(Config{
		DisableGoCollector:      true,
		DisableProcessCollector: true,
	})

	suite.NoError(err)
	suite.NotNil(g)
	suite.NotNil(r)

	suite.NoError(
		r.Register(prometheus.NewGoCollector()),
	)

	suite.NoError(
		r.Register(prometheus.NewProcessCollector(
			prometheus.ProcessCollectorOpts{},
		)),
	)
}

func TestNew(t *testing.T) {
	suite.Run(t, new(NewTestSuite))
}
