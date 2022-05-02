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

package touchhttp

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/suite"
)

type ErrorPrinterTestSuite struct {
	suite.Suite

	output strings.Builder
}

// Printf lets this suite be seen as an fx.Printer.  Output is redirected
// to an internal buffer for verification.
func (suite *ErrorPrinterTestSuite) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&suite.output, format, args...)
}

func (suite *ErrorPrinterTestSuite) setupErrorPrinter() ErrorPrinter {
	suite.output.Reset()
	return ErrorPrinter{
		Printer: suite,
	}
}

func (suite *ErrorPrinterTestSuite) TestNoArguments() {
	suite.output.Reset()
	ep := suite.setupErrorPrinter()
	ep.Println()
	suite.Zero(suite.output.Len())
}

func (suite *ErrorPrinterTestSuite) TestPrintln() {
	testCases := [][]interface{}{
		[]interface{}{"test"},
		[]interface{}{"test", 123, 18 * time.Millisecond},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("args=%d", len(testCase)), func() {
			ep := suite.setupErrorPrinter()

			ep.Println(testCase...)
			fields := strings.Split(suite.output.String(), " ")
			suite.Equal(len(testCase), len(fields))
		})
	}
}

func TestErrorPrinter(t *testing.T) {
	suite.Run(t, new(ErrorPrinterTestSuite))
}

type NewHandlerOptsTestSuite struct {
	suite.Suite

	output strings.Builder
}

var _ suite.SetupTestSuite = (*NewHandlerOptsTestSuite)(nil)

func (suite *NewHandlerOptsTestSuite) SetupTest() {
	suite.output.Reset()
}

// Printf lets this suite be seen as an fx.Printer.  Output is redirected
// to an internal buffer for verification.
func (suite *NewHandlerOptsTestSuite) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&suite.output, format, args...)
}

func (suite *NewHandlerOptsTestSuite) TestDefaults() {
	ho, err := NewHandlerOpts(Config{}, nil, nil)
	suite.NoError(err)
	suite.Nil(ho.ErrorLog)
	suite.Equal(promhttp.HTTPErrorOnError, ho.ErrorHandling)
	suite.False(ho.DisableCompression)
	suite.Zero(ho.MaxRequestsInFlight)
	suite.Zero(ho.Timeout)
	suite.False(ho.EnableOpenMetrics)
	suite.Nil(ho.Registry)
}

func (suite *NewHandlerOptsTestSuite) TestCustom() {
	r := prometheus.NewRegistry()
	ho, err := NewHandlerOpts(
		Config{
			ErrorHandling:       ContinueOnError,
			DisableCompression:  true,
			MaxRequestsInFlight: 20,
			Timeout:             17 * time.Hour,
			EnableOpenMetrics:   true,
		},
		suite,
		r,
	)

	suite.NoError(err)
	suite.Equal(promhttp.ContinueOnError, ho.ErrorHandling)
	suite.True(ho.DisableCompression)
	suite.Equal(20, ho.MaxRequestsInFlight)
	suite.Equal(17*time.Hour, ho.Timeout)
	suite.True(ho.EnableOpenMetrics)
	suite.Equal(r, ho.Registry)

	suite.Require().NotNil(ho.ErrorLog)
	ho.ErrorLog.Println("test", 123)
	suite.NotZero(suite.output.Len())
}

func (suite *NewHandlerOptsTestSuite) TestErrorHandlingValues() {
	suite.Run("Invalid", func() {
		_, err := NewHandlerOpts(
			Config{
				ErrorHandling: "this is an invalid value",
			}, nil, nil,
		)

		suite.Error(err)

		var actualErr *InvalidErrorHandlingError
		suite.Require().True(errors.As(err, &actualErr))
		suite.NotEmpty(actualErr.Error())
	})

	suite.Run("Valid", func() {
		testCases := []struct {
			value    string
			expected promhttp.HandlerErrorHandling
		}{
			{value: "", expected: promhttp.HTTPErrorOnError},
			{value: HTTPErrorOnError, expected: promhttp.HTTPErrorOnError},
			{value: ContinueOnError, expected: promhttp.ContinueOnError},
			{value: PanicOnError, expected: promhttp.PanicOnError},
		}

		for _, testCase := range testCases {
			suite.Run(fmt.Sprintf("'%s'", testCase.value), func() {
				ho, err := NewHandlerOpts(
					Config{
						ErrorHandling: testCase.value,
					}, nil, nil,
				)

				suite.NoError(err)
				suite.Equal(testCase.expected, ho.ErrorHandling)
			})
		}
	})
}

func TestNewHandlerOpts(t *testing.T) {
	suite.Run(t, new(NewHandlerOptsTestSuite))
}
