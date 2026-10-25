// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/fx"
)

func ExampleFactory() {
	var f *Factory
	_ = fx.New(
		fx.NopLogger,
		Provide(),
		fx.Populate(&f),
	)

	c, err := f.NewCounterVec(
		prometheus.CounterOpts{
			Name: "example",
			Help: "here is a lovely example",
		}, "label",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	c.WithLabelValues("value").Inc()
	fmt.Println(testutil.ToFloat64(c))

	// Output:
	// 1
}
