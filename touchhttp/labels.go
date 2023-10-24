// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchhttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// CodeLabel is the metric label containing the HTTP response code.
	CodeLabel = "code"

	// MethodLabel is the metric label containing the HTTP request's method.
	MethodLabel = "method"

	// ServerLabel is the canonical metric label name containing the name of the HTTP server.
	// This label is not automatically supplied.
	ServerLabel = "server"

	// ClientLabel is the canonical metric label name containing the name of the HTTP client.
	// This label is not automatically supplied.
	ClientLabel = "client"

	// MethodUnrecognized is used when an HTTP method is not one of the
	// standard methods, as enumerated in the net/http package.
	MethodUnrecognized = "UNRECOGNIZED"
)

var (
	// recognizedMethods is the set of HTTP methods that are defined by the spec(s).
	// Any method not in this map gets recorded as MethodUnknown.
	recognizedMethods = map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodPatch:   true,
		http.MethodDelete:  true,
		http.MethodConnect: true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
	}

	statusUnknown = "-1"
	statusOK      = "200"

	// preformattedCodes is used to avoid extra allocation overhead for
	// common status codes
	preformattedCodes = map[int]string{
		200: statusOK,
		201: "201",
		202: "202",
		204: "204",
		400: "400",
		401: "401",
		403: "403",
		404: "404",
		405: "405",
		429: "429",
		500: "500",
		502: "502",
		503: "503",
		504: "504",
	}
)

// formatCode is an efficient, zero-copy formatter for 3-digit HTTP response codes.
// this function avoids the general stdlib in favor of an unwound loop specific to
// 3-digit integers in the valid range of HTTP status codes.  It also uses preallocated
// strings for common status codes.
func formatCode(v int) string {
	switch {
	case v == 0:
		// if an http.Handler never calls WriteHeader, the status code will be 0
		// but a 200 is assumed
		return statusOK

	case v < 100:
		return statusUnknown

	case v > 599:
		return statusUnknown
	}

	if vv, ok := preformattedCodes[v]; ok {
		return vv
	}

	var code [3]byte
	code[2] = '0' + byte(v%10)
	v /= 10
	code[1] = '0' + byte(v%10)
	v /= 10
	code[0] = '0' + byte(v%10)

	return string(code[:])
}

// formatMethod ensures that its argument is a valid HTTP method and
// returns that value.  if it's not valid, methodUnrecognized is returned.
func formatMethod(v string) string {
	if recognizedMethods[v] {
		return v
	}

	return MethodUnrecognized
}

// Labels is a convenient extension for a prometheus.Labels that
// adds support for the reserved and de facto labels in this package.
//
// The zero value for this map is usable with any of its methods.
type Labels prometheus.Labels

func (l *Labels) set(k, v string) {
	if *l == nil {
		*l = make(Labels)
	}

	(*l)[k] = v
}

// SetCode updates this set of Labels with the given HTTP status code
// Passing a value outside the range of valid HTTP response code values
// (e.g. zero) means http.StatusOK.
func (l *Labels) SetCode(v int) {
	l.set(CodeLabel, formatCode(v))
}

// SetMethod updates this set of Labels with the given HTTP method.
// Any unrecognized methods result in MethodUnrecognized.
func (l *Labels) SetMethod(v string) {
	l.set(MethodLabel, formatMethod(v))
}

// SetServer updates this set of Labels with the given server name.
func (l *Labels) SetServer(v string) {
	l.set(ServerLabel, v)
}

// SetClient updates this set of Labels with the given client name.
func (l *Labels) SetClient(v string) {
	l.set(ClientLabel, v)
}

// NewLabels creates a touchhttp Labels with code and method set appropriately.
func NewLabels(code int, method string) Labels {
	return Labels{
		CodeLabel:   formatCode(code),
		MethodLabel: formatMethod(method),
	}
}
