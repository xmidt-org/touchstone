// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"bytes"
	"io"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/suite"
)

type NewUntypedFuncSuite struct {
	suite.Suite
}

func (suite *NewUntypedFuncSuite) newRegistry() *prometheus.Registry {
	return prometheus.NewPedanticRegistry()
}

func (suite *NewUntypedFuncSuite) newUntypedOpts() prometheus.UntypedOpts {
	return prometheus.UntypedOpts{
		Name: "untyped",
		Help: "untyped",
	}
}

func (suite *NewUntypedFuncSuite) setupExpectation(v float64) io.Reader {
	registry := suite.newRegistry()

	suite.Require().NoError(
		registry.Register(
			prometheus.NewUntypedFunc(
				suite.newUntypedOpts(),
				func() float64 { return v },
			),
		),
	)

	families, err := registry.Gather()
	suite.Require().NoError(err)

	var (
		output  = new(bytes.Buffer)
		encoder = expfmt.NewEncoder(output, expfmt.FmtText)
	)

	if c, ok := encoder.(io.Closer); ok {
		// the docs say we should always call close, so ...
		defer c.Close()
	}

	for _, fam := range families {
		suite.Require().NoError(
			encoder.Encode(fam),
		)
	}

	return output
}

func (suite *NewUntypedFuncSuite) TestSupportedType() {
	testCases := []struct {
		description string
		f           interface{}
		expected    float64
	}{
		{
			description: "uint8",
			f:           func() uint8 { return 12 },
			expected:    12.0,
		},
		{
			description: "uint16",
			f:           func() uint16 { return 8264 },
			expected:    8264.0,
		},
		{
			description: "uint32",
			f:           func() uint32 { return 348 },
			expected:    348.0,
		},
		{
			description: "uint64",
			f:           func() uint64 { return 92642 },
			expected:    92642.0,
		},
		{
			description: "uint",
			f:           func() uint { return 264 },
			expected:    264.0,
		},
		{
			description: "int8",
			f:           func() int8 { return 12 },
			expected:    12.0,
		},
		{
			description: "int16",
			f:           func() int16 { return 8264 },
			expected:    8264.0,
		},
		{
			description: "int32",
			f:           func() int32 { return 348 },
			expected:    348.0,
		},
		{
			description: "int64",
			f:           func() int64 { return 92642 },
			expected:    92642.0,
		},
		{
			description: "int",
			f:           func() int { return 264 },
			expected:    264.0,
		},
		{
			description: "float32",
			f:           func() float32 { return -462.0 },
			expected:    -462.0,
		},
		{
			description: "float64",
			f:           func() float64 { return 1234.0 },
			expected:    1234.0,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.description, func() {
			m, err := NewUntypedFunc(suite.newUntypedOpts(), testCase.f)
			suite.Require().NoError(err)

			actual := suite.newRegistry()
			suite.Require().NoError(actual.Register(m))

			expected := suite.setupExpectation(testCase.expected)
			suite.NoError(
				testutil.GatherAndCompare(actual, expected),
			)
		})
	}
}

func (suite *NewUntypedFuncSuite) TestUnsupportedType() {
	testCases := []struct {
		description string
		f           interface{}
	}{
		{
			description: "nil",
			f:           nil,
		},
		{
			description: "not a function",
			f:           77,
		},
		{
			description: "bad return value",
			f:           func() string { return "oh no!" },
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.description, func() {
			_, err := NewUntypedFunc(suite.newUntypedOpts(), testCase.f)
			suite.Error(err)
		})
	}
}

func TestNewUntypedFunc(t *testing.T) {
	suite.Run(t, new(NewUntypedFuncSuite))
}
