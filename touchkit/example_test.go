// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchkit_test

import (
	"fmt"
	"strings"

	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/xmidt-org/touchstone"
	"github.com/xmidt-org/touchstone/touchkit"
	"go.uber.org/fx"
)

// Example shows how code written against go-kit's metrics interfaces can
// receive prometheus-backed metrics through dependency injection.  Each
// component is named after its metric, so a struct with fx.In can select
// the metric it wants by name.
func Example() {
	type In struct {
		fx.In

		Requests metrics.Counter   `name:"requests"`
		Latency  metrics.Histogram `name:"latency"`
	}

	var g prometheus.Gatherer
	app := fx.New(
		fx.NopLogger,
		touchstone.Provide(),
		touchkit.Counter(
			prometheus.CounterOpts{Name: "requests", Help: "the total number of requests"},
			"code",
		),
		touchkit.Histogram(
			prometheus.HistogramOpts{Name: "latency", Help: "latency in milliseconds", Buckets: []float64{10, 100}},
		),
		fx.Populate(&g),
		fx.Invoke(func(in In) {
			in.Requests.With("code", "200").Add(3)
			in.Latency.Observe(42)
		}),
	)

	if err := app.Err(); err != nil {
		fmt.Println(err)
		return
	}

	expected := `
# HELP requests the total number of requests
# TYPE requests counter
requests{code="200"} 3
`

	fmt.Println(testutil.GatherAndCompare(g, strings.NewReader(expected), "requests"))

	count, err := testutil.GatherAndCount(g, "latency")
	fmt.Println(count, err)

	// Output:
	// <nil>
	// 1 <nil>
}
