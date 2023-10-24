// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// NewUntypedFunc is a variant of prometheus.NewUntypedFunc that allows the
// function to have a more flexible signature.  The supplied function f must accept
// no arguments and must return exactly (1) value that is any scalar numeric type.
// The complex types are not supported.
//
// In particular, this function is useful when f has the signature func() int.  This
// is the common case for things like queue depth, length of a data structure, etc.
//
// If f is not a function or is a function with an unsupported signature,
// an error is returned.
func NewUntypedFunc(opts prometheus.UntypedOpts, f interface{}) (uf prometheus.UntypedFunc, err error) {
	var untyped func() float64
	switch fn := f.(type) {
	case func() uint8: // handles byte
		untyped = func() float64 { return float64(fn()) }

	case func() uint16:
		untyped = func() float64 { return float64(fn()) }

	case func() uint32:
		untyped = func() float64 { return float64(fn()) }

	case func() uint64:
		untyped = func() float64 { return float64(fn()) }

	case func() uint:
		untyped = func() float64 { return float64(fn()) }

	case func() int8:
		untyped = func() float64 { return float64(fn()) }

	case func() int16:
		untyped = func() float64 { return float64(fn()) }

	case func() int32: // handles rune
		untyped = func() float64 { return float64(fn()) }

	case func() int64:
		untyped = func() float64 { return float64(fn()) }

	case func() int:
		untyped = func() float64 { return float64(fn()) }

	case func() float32:
		untyped = func() float64 { return float64(fn()) }

	case func() float64:
		untyped = fn

	default:
		err = fmt.Errorf(
			"%T is not a function with the signature func() N, where N is a numeric type",
			f,
		)
	}

	if err == nil {
		uf = prometheus.NewUntypedFunc(opts, untyped)
	}

	return
}
