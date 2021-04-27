package touchhttp

import (
	"io"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
)

const (
	testServerName = "test"
	testClientName = "test"

	testDuration time.Duration = 151 * time.Millisecond
)

// BundleTestSuite is all the common functionality for testing metric bundles.
type BundleTestSuite struct {
	suite.Suite

	// now is a known start time for all clocks
	now time.Time
}

var _ suite.SetupTestSuite = (*BundleTestSuite)(nil)

func (suite *BundleTestSuite) SetupTest() {
	suite.now = time.Now()
}

func (suite *BundleTestSuite) Printf(format string, args ...interface{}) {
	suite.T().Logf(format, args...)
}

// newRegistry constructs a registry and its wrapping factory with a prebuilt
// environment for testing.
func (suite *BundleTestSuite) newRegistry() (*prometheus.Registry, *touchstone.Factory) {
	r := prometheus.NewPedanticRegistry()
	f := touchstone.NewFactory(touchstone.Config{}, suite, r)
	return r, f
}

// encode writes a gatherer's state to the writer using the testutil package.
// This enables comparing the state of two registries.
func (suite *BundleTestSuite) encode(o io.Writer, g prometheus.Gatherer) {
	raw, err := g.Gather()
	suite.Require().NoError(err)

	encoder := expfmt.NewEncoder(o, expfmt.FmtText)
	if closer, ok := encoder.(expfmt.Closer); ok {
		defer closer.Close()
	}

	for _, mf := range raw {
		suite.Require().NoError(encoder.Encode(mf))
	}
}

// newClock creates a now implementation that produces a known duration.
func (suite *BundleTestSuite) newClock() clock {
	return clock{
		start:    suite.now,
		duration: testDuration,
	}
}
