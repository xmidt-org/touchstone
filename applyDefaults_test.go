// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type ApplyDefaultsTestSuite struct {
	suite.Suite
}

func (suite *ApplyDefaultsTestSuite) TestSrc() {
	suite.Run("NilInterface", func() {
		actual := prometheus.CounterOpts{
			Name: "this should be unchanged",
		}

		suite.T().Log("a nil src interface should be a noop")
		result := ApplyDefaults(&actual, nil)
		suite.Equal(
			prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			actual,
		)

		suite.Equal(
			&prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			result,
		)
	})

	suite.Run("NilPointerToStruct", func() {
		actual := prometheus.CounterOpts{
			Name: "this should be unchanged",
		}

		suite.T().Log("a nil src pointer to a struct should be a noop")
		result := ApplyDefaults(&actual, (*prometheus.CounterOpts)(nil))
		suite.Equal(
			prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			actual,
		)

		suite.Equal(
			&prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			result,
		)
	})

	suite.Run("WrongTypeOfPointer", func() {
		actual := prometheus.CounterOpts{
			Name: "this should be unchanged",
		}

		suite.T().Log("a src pointer to anything other than a struct should panic")
		suite.Panics(func() {
			ApplyDefaults(&actual, (*int)(nil))
		})

		suite.Equal(
			prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			actual,
		)
	})

	suite.Run("WrongType", func() {
		actual := prometheus.CounterOpts{
			Name: "this should be unchanged",
		}

		suite.T().Log("a non-struct src should panic")
		suite.Panics(func() {
			ApplyDefaults(&actual, 123)
		})

		suite.Equal(
			prometheus.CounterOpts{
				Name: "this should be unchanged",
			},
			actual,
		)
	})
}

func (suite *ApplyDefaultsTestSuite) TestDst() {
	suite.Run("NilInterface", func() {
		suite.T().Log("a nil dst interface should panic")
		var defaults prometheus.CounterOpts
		suite.Panics(func() {
			ApplyDefaults(nil, defaults)
		})
	})

	suite.Run("NilPointerToStruct", func() {
		suite.T().Log("a nil dst pointer should result in a new struct")
		defaults := prometheus.CounterOpts{
			Name: "default",
			Help: "default",
		}

		result := ApplyDefaults((*prometheus.CounterOpts)(nil), defaults)
		suite.Equal(
			&prometheus.CounterOpts{
				Name: "default",
				Help: "default",
			},
			result,
		)
	})

	suite.Run("WrongTypeOfPointer", func() {
		suite.T().Log("a dst pointer to anything other than a struct should panic")
		var defaults prometheus.CounterOpts
		var wrong int
		suite.Panics(func() {
			ApplyDefaults(&wrong, defaults)
		})
	})

	suite.Run("NotAPointer", func() {
		suite.T().Log("a dst value that is not a pointer should panic")
		var defaults prometheus.CounterOpts
		var wrong int
		suite.Panics(func() {
			ApplyDefaults(wrong, defaults)
		})
	})
}

func (suite *ApplyDefaultsTestSuite) TestApply() {
	// used for anonymous field testing
	type Name string

	// nonstandard dst struct
	type TestDst struct {
		Name
		Namespace  string
		Help       int
		unexported string
	}

	testCases := []struct {
		dst      interface{}
		src      interface{}
		expected interface{}
	}{
		{
			dst: &prometheus.CounterOpts{
				Name: "test_counter",
			},
			// struct (i.e. not a pointer)
			src: prometheus.Opts{
				Namespace: "default",
				Help:      "default",
			},
			expected: &prometheus.CounterOpts{
				Namespace: "default",
				Name:      "test_counter",
				Help:      "default",
			},
		},
		{
			dst: prometheus.CounterOpts{
				Name: "test_counter",
				Help: "pass_by_value",
			},
			// struct (i.e. not a pointer)
			src: prometheus.Opts{
				Namespace: "default",
				Help:      "default",
			},
			expected: &prometheus.CounterOpts{
				Namespace: "default",
				Name:      "test_counter",
				Help:      "pass_by_value",
			},
		},
		{
			dst: &prometheus.CounterOpts{
				Name: "test_counter",
			},
			// pointer to struct
			src: &prometheus.Opts{
				Namespace: "default",
				Help:      "default",
			},
			expected: &prometheus.CounterOpts{
				Namespace: "default",
				Name:      "test_counter",
				Help:      "default",
			},
		},
		{
			dst: &prometheus.HistogramOpts{
				Help: "test_help",
			},
			src: struct {
				Name        // should be skipped: anonymous
				Help    int // should be skipped: wrong type
				Buckets []float64
			}{
				Name:    "should be skipped",
				Help:    123,
				Buckets: []float64{1, 10, 100, 1000},
			},
			expected: &prometheus.HistogramOpts{
				Help:    "test_help",
				Buckets: []float64{1, 10, 100, 1000},
			},
		},
		{
			dst: &TestDst{
				Name:       "should be ignored",
				Help:       123,
				unexported: "unexported",
			},
			src: prometheus.Opts{
				Namespace: "default",
				Name:      "this should be skipped",
				Help:      "this should be skipped",
			},
			expected: &TestDst{
				Namespace:  "default",
				Name:       "should be ignored",
				Help:       123,
				unexported: "unexported",
			},
		},
	}

	for i, testCase := range testCases {
		suite.Run(strconv.Itoa(i), func() {
			result := ApplyDefaults(testCase.dst, testCase.src)
			suite.Equal(testCase.expected, result)
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	suite.Run(t, new(ApplyDefaultsTestSuite))
}
