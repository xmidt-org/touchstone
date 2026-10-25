// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchbundle_test

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/touchstone/touchbundle"
	"go.uber.org/fx"
)

// Metrics is a bundle: a plain struct whose fields describe the metrics an
// application needs.  Struct tags supply the metric options, and any field
// name without an explicit name tag is converted to snake case.
type Metrics struct {
	// Requests becomes a counter named "requests".
	Requests prometheus.Counter `help:"the total number of requests"`

	// RequestErrors becomes a counter vector with a single label.
	RequestErrors *prometheus.CounterVec `help:"errors by type" labelNames:"type"`

	// Latency is an observer, so the type tag chooses between a histogram and a summary.
	Latency prometheus.Observer `type:"histogram" help:"latency in milliseconds" buckets:"10, 50, 100"`

	// Name is not a metric and is ignored.
	Name string `touchstone:"-"`
}

// Example shows how a bundle participates in dependency injection.  Provide
// creates a new instance of the prototype, populates its metric fields, and
// registers each metric with the application's prometheus registry.
func Example() {
	var m Metrics
	app := fx.New(
		fx.NopLogger,
		touchstone.Provide(),
		touchbundle.Provide(Metrics{}),
		fx.Populate(&m),
	)

	if err := app.Err(); err != nil {
		fmt.Println(err)
		return
	}

	m.Requests.Inc()
	m.RequestErrors.WithLabelValues("timeout").Add(2)
	m.Latency.Observe(42)

	fmt.Println("requests", testutil.ToFloat64(m.Requests))
	fmt.Println("request_errors", testutil.ToFloat64(m.RequestErrors))

	// Output:
	// requests 1
	// request_errors 2
}

// ExamplePopulate shows how to fill a bundle without an fx.App, given a
// touchstone.Factory.  The Factory's default namespace is applied to every
// metric in the bundle.
func ExamplePopulate() {
	cfg := touchstone.Config{DefaultNamespace: "example"}
	g, r, err := touchstone.New(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	var m Metrics
	if err := touchbundle.Populate(touchstone.NewFactory(cfg, nil, r), &m); err != nil {
		fmt.Println(err)
		return
	}

	m.Requests.Add(3)

	count, err := testutil.GatherAndCount(g, "example_requests")
	fmt.Println(count, err)
	fmt.Println(testutil.ToFloat64(m.Requests))

	// Output:
	// 1 <nil>
	// 3
}
