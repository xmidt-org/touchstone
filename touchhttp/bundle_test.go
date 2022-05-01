package touchhttp

import (
	"time"

	"github.com/stretchr/testify/suite"
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
