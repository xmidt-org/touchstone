// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchbundle

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		identifier string
		expected   string
	}{
		{identifier: "", expected: ""},
		{identifier: "simple", expected: "simple"},
		{identifier: "Simple", expected: "simple"},
		{identifier: "MyAppCounter", expected: "my_app_counter"},
		{identifier: "myAppCounter", expected: "my_app_counter"},

		// runs of capitals are treated as a single token
		{identifier: "HTTP", expected: "http"},
		{identifier: "HTTPServer", expected: "http_server"},
		{identifier: "ServerHTTP", expected: "server_http"},
		{identifier: "ABCdef", expected: "ab_cdef"},

		// explicit separators are preserved, but not doubled or at the ends
		{identifier: "_Leading", expected: "leading"},
		{identifier: "Trailing_", expected: "trailing"},
		{identifier: "Mixed_Separator", expected: "mixed_separator"},
		{identifier: "Double__Separator", expected: "double_separator"},
		{identifier: "already_snake", expected: "already_snake"},

		// digits stay attached to the current token
		{identifier: "Count2Total", expected: "count2_total"},
		{identifier: "V2", expected: "v2"},

		// title case letters start a token just like upper case letters
		{identifier: "ǅemal", expected: "ǆemal"},
	}

	for _, tc := range tests {
		t.Run(tc.identifier, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, toSnakeCase(tc.identifier))
		})
	}
}

func TestMetricName(t *testing.T) {
	t.Parallel()

	type bundle struct {
		SomethingCount int
		MyGauge        int `name:"prefix_*"`
		Suffixed       int `name:"*_total"`
		AnotherCounter int `name:"custom_name"`
		EmptyTag       int `name:""`
	}

	tests := []struct {
		field    string
		expected string
	}{
		{field: "SomethingCount", expected: "something_count"},
		{field: "MyGauge", expected: "prefix_my_gauge"},
		{field: "Suffixed", expected: "suffixed_total"},
		{field: "AnotherCounter", expected: "custom_name"},
		{field: "EmptyTag", expected: "empty_tag"},
	}

	for _, tc := range tests {
		t.Run(tc.field, func(t *testing.T) {
			t.Parallel()

			f, ok := reflect.TypeFor[bundle]().FieldByName(tc.field)
			assert.True(t, ok)
			assert.Equal(t, tc.expected, MetricName(f))
		})
	}
}
