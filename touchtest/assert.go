// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchtest

import (
	"bytes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Assertions is a set of test verifications for metrics.  Principally,
// this involves comparisons against an expected Gatherer.
type Assertions struct {
	buffer bytes.Buffer
	names  map[string]bool

	assert  *assert.Assertions
	require *require.Assertions
}

// New creates an Assertions for the given testing environment.
func New(t require.TestingT) *Assertions {
	return &Assertions{
		assert:  assert.New(t),
		require: require.New(t),
	}
}

// NewSuite creates an Assertions for the given testing suite.
func NewSuite(s suite.TestingSuite) *Assertions {
	return New(s.T())
}

// Expect loads a given set of metrics into this set of assertions.
// Multiple assertions can be made against this expected set without
// reencoding.
//
// If any errors occur while encoding the given expected Gatherer,
// the enclosing test is failed.
func (a *Assertions) Expect(g prometheus.Gatherer) *Assertions {
	raw, err := g.Gather()
	if err == nil {
		a.buffer.Reset()
		a.names = make(map[string]bool)
		enc := expfmt.NewEncoder(&a.buffer, expfmt.FmtText)

		if closer, ok := enc.(expfmt.Closer); ok {
			defer closer.Close()
		}

		for i := 0; err == nil && i < len(raw); i++ {
			mf := raw[i]
			if err = enc.Encode(mf); err == nil && mf.Name != nil {
				a.names[*mf.Name] = true
			}
		}
	}

	a.require.NoError(err, "Failed to set expected Gatherer")
	return a
}

// GatherAndCompare compares the actual Gatherer against the current expectation
// previously set with Expect.  This method fails the test on any error, then returns false.
// This method returns true if the expectation was met.
//
// Use this method to run an assertion against an entire prometheus registry.
func (a *Assertions) GatherAndCompare(actual prometheus.Gatherer, metricNames ...string) bool {
	err := testutil.GatherAndCompare(
		actual,
		bytes.NewReader(a.buffer.Bytes()),
		metricNames...,
	)

	return a.assert.NoErrorf(
		err,
		"Failed to match expected metrics: %s",
		metricNames,
	)
}

// CollectAndCompare compares the actual collector against the current expectation
// previously set with Expect.  This method fails the test on any error, then returns false.
// This method returns true if the expectation was met.
//
// Use this method to run an assertion against a single metric that optionally has
// multiple submetrics.
func (a *Assertions) CollectAndCompare(actual prometheus.Collector, metricNames ...string) bool {
	err := testutil.CollectAndCompare(
		actual,
		bytes.NewReader(a.buffer.Bytes()),
		metricNames...,
	)

	return a.assert.NoErrorf(
		err,
		"Failed to match expected metrics", "names: %s",
		metricNames,
	)
}

// Registered asserts that the given metric names are present in the current expectation
// previously set with Expect.
func (a *Assertions) Registered(metricNames ...string) bool {
	passed := true
	for _, n := range metricNames {
		passed = a.assert.Truef(a.names[n], "Metric SHOULD BE registered: %s", n) && passed
	}

	return passed
}

// NotRegistered asserts that the given metric names are absent in the current expectation
// previously set with Expect.
func (a *Assertions) NotRegistered(metricNames ...string) bool {
	passed := true
	for _, n := range metricNames {
		passed = a.assert.Falsef(a.names[n], "Metric SHOULD NOT BE registered: %s", n) && passed
	}

	return passed
}
