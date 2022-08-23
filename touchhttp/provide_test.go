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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type ProvideTestSuite struct {
	suite.Suite
}

// assertMetricExists verifies that a metric exists by simply attempting to
// register a counter with the given name.
func (suite *ProvideTestSuite) assertMetricExists(r prometheus.Registerer, name, msg string) bool {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{Name: name},
	)

	return suite.Error(r.Register(c), msg)
}

// assertMetricNotExists verifies that a metric exists by simply attempting to
// register a counter with the given name.
func (suite *ProvideTestSuite) assertMetricNotExists(r prometheus.Registerer, name, msg string) bool {
	c := prometheus.NewCounter(
		prometheus.CounterOpts{Name: name},
	)

	return suite.NoError(r.Register(c), msg)
}

func (suite *ProvideTestSuite) TestDefaults() {
	var (
		ho promhttp.HandlerOpts
		h  Handler
		r  prometheus.Registerer

		app = fxtest.New(
			suite.T(),
			touchstone.Provide(),
			Provide(),
			fx.Populate(&ho, &h, &r),
		)
	)

	suite.NoError(app.Err())
	app.RequireStart()

	suite.assertMetricNotExists(
		r,
		"promhttp_metric_handler_requests_total",
		"Handler should not have been instrumented",
	)

	app.RequireStop()
}

func (suite *ProvideTestSuite) TestInstrumentMetricHandler() {
	var (
		ho promhttp.HandlerOpts
		h  Handler
		r  prometheus.Registerer

		app = fxtest.New(
			suite.T(),
			fx.Supply(
				Config{
					InstrumentMetricHandler: true,
				},
			),
			touchstone.Provide(),
			Provide(),
			fx.Populate(&ho, &h, &r),
		)
	)

	suite.NoError(app.Err())
	app.RequireStart()

	suite.assertMetricExists(
		r,
		"promhttp_metric_handler_requests_total",
		"Handler should have been instrumented",
	)

	app.RequireStop()
}

func TestProvide(t *testing.T) {
	suite.Run(t, new(ProvideTestSuite))
}
