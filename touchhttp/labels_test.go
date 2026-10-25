// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     int
		expected string
	}{
		// an unset status code means the handler never called WriteHeader
		{code: 0, expected: "200"},

		// out of range values
		{code: -1, expected: statusUnknown},
		{code: 99, expected: statusUnknown},
		{code: 600, expected: statusUnknown},
		{code: 1000, expected: statusUnknown},

		// preformatted values
		{code: 200, expected: "200"},
		{code: 404, expected: "404"},
		{code: 503, expected: "503"},

		// values that must be computed
		{code: 100, expected: "100"},
		{code: 203, expected: "203"},
		{code: 418, expected: "418"},
		{code: 599, expected: "599"},
	}

	for _, tc := range tests {
		t.Run(strconv.Itoa(tc.code), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, formatCode(tc.code))
		})
	}
}

// TestFormatCodeMatchesStrconv verifies the hand-unrolled formatter against
// the standard library for every valid HTTP status code.
func TestFormatCodeMatchesStrconv(t *testing.T) {
	t.Parallel()

	for v := 100; v <= 599; v++ {
		assert.Equal(t, strconv.Itoa(v), formatCode(v), "code %d", v)
	}
}

func TestFormatMethod(t *testing.T) {
	t.Parallel()

	t.Run("Recognized", func(t *testing.T) {
		t.Parallel()
		for method := range recognizedMethods {
			assert.Equal(t, method, formatMethod(method))
		}
	})

	t.Run("Unrecognized", func(t *testing.T) {
		t.Parallel()
		for _, method := range []string{"", "BREW", "get", "Post"} {
			assert.Equal(t, MethodUnrecognized, formatMethod(method), "method %q", method)
		}
	})
}

func TestLabels(t *testing.T) {
	t.Parallel()

	t.Run("ZeroValue", func(t *testing.T) {
		t.Parallel()

		var l Labels
		l.SetCode(0)
		l.SetMethod(http.MethodGet)
		l.SetServer("main")
		l.SetClient("api")

		assert.Equal(t, Labels{
			CodeLabel:   "200",
			MethodLabel: http.MethodGet,
			ServerLabel: "main",
			ClientLabel: "api",
		}, l)
	})

	t.Run("Existing", func(t *testing.T) {
		t.Parallel()

		l := Labels{"custom": "value"}
		l.SetCode(http.StatusNotFound)
		l.SetMethod("BREW")

		assert.Equal(t, Labels{
			"custom":    "value",
			CodeLabel:   "404",
			MethodLabel: MethodUnrecognized,
		}, l)
	})

	t.Run("Overwrite", func(t *testing.T) {
		t.Parallel()

		l := NewLabels(http.StatusOK, http.MethodGet)
		l.SetCode(http.StatusInternalServerError)
		l.SetMethod(http.MethodDelete)

		assert.Equal(t, Labels{
			CodeLabel:   "500",
			MethodLabel: http.MethodDelete,
		}, l)
	})
}

func TestNewLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     int
		method   string
		expected Labels
	}{
		{
			name:     "Valid",
			code:     http.StatusCreated,
			method:   http.MethodPost,
			expected: Labels{CodeLabel: "201", MethodLabel: http.MethodPost},
		},
		{
			name:     "Invalid",
			code:     -1,
			method:   "nope",
			expected: Labels{CodeLabel: statusUnknown, MethodLabel: MethodUnrecognized},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, NewLabels(tc.code, tc.method))
		})
	}
}
