package touchbundle

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	testCases := []struct {
		identifier string
		expected   string
	}{
		{
			// empty identifier should result in the empty string
		},
		{
			identifier: "Simple",
			expected:   "simple",
		},
		{
			identifier: "RequestCount",
			expected:   "request_count",
		},
		{
			identifier: "RequestURI",
			expected:   "request_uri",
		},
		{
			identifier: "INITIALCAPS",
			expected:   "initialcaps",
		},
		{
			identifier: "Preserve_underscore",
			expected:   "preserve_underscore",
		},
		{
			identifier: "SomethingABCDoit",
			expected:   "something_abc_doit",
		},
		{
			identifier: "AComplex00Identifier",
			expected:   "a_complex00_identifier",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.identifier, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(testCase.expected, toSnakeCase(testCase.identifier))
		})
	}
}

func TestMetricName(t *testing.T) {
	testCases := []struct {
		description        string
		fieldName          string
		fieldTag           reflect.StructTag
		expectedMetricName string
	}{
		{
			description:        "simple",
			fieldName:          "RequestCount",
			expectedMetricName: "request_count",
		},
		{
			description:        "custom name",
			fieldName:          "RequestCount",
			fieldTag:           `name:"custom_name"`,
			expectedMetricName: "custom_name",
		},
		{
			description:        "prefix",
			fieldName:          "RequestCount",
			fieldTag:           `name:"prefix_*"`,
			expectedMetricName: "prefix_request_count",
		},
		{
			description:        "suffix",
			fieldName:          "RequestCount",
			fieldTag:           `name:"*_suffix"`,
			expectedMetricName: "request_count_suffix",
		},
		{
			description:        "internal",
			fieldName:          "RequestCount",
			fieldTag:           `name:"a_*_b"`,
			expectedMetricName: "a_request_count_b",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(
				t,
				testCase.expectedMetricName,
				MetricName(reflect.StructField{
					Name: testCase.fieldName,
					Tag:  testCase.fieldTag,
				}),
			)
		})
	}
}
