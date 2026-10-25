// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchtest_test

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/touchstone/touchtest"
)

// exampleT stands in for a *testing.T so the example can run outside a
// test.  In real tests, pass the *testing.T directly to touchtest.New.
type exampleT struct{}

func (exampleT) Errorf(string, ...any) { fmt.Println("assertion failed") }

func (exampleT) FailNow() {}

func newCounter(r prometheus.Registerer, value float64) prometheus.Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests",
		Help: "the total number of requests",
	})

	r.MustRegister(c)
	c.Add(value)
	return c
}

// Example shows how to assert on metrics by describing the expected state
// in a separate registry.  Expect loads that registry, after which the
// comparison methods check an actual Gatherer or Collector against it.
func Example() {
	expected := prometheus.NewRegistry()
	newCounter(expected, 3)

	actual := prometheus.NewRegistry()
	counter := newCounter(actual, 3)

	a := touchtest.New(exampleT{})
	a.Expect(expected)

	fmt.Println(a.Registered("requests"))
	fmt.Println(a.NotRegistered("other"))
	fmt.Println(a.GatherAndCompare(actual, "requests"))
	fmt.Println(a.CollectAndCompare(counter, "requests"))

	// a change to the actual metric no longer matches the expectation
	counter.Inc()
	fmt.Println(a.CollectAndCompare(counter, "requests"))

	// Output:
	// true
	// true
	// true
	// true
	// assertion failed
	// false
}
