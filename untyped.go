/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
// If f is not a function or is a function with a supported signature,
// this function panics.
func NewUntypedFunc(opts prometheus.UntypedOpts, f interface{}) prometheus.UntypedFunc {
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
		panic(
			fmt.Errorf(
				"%T is not a function with the signature func() N, where N is a numeric type",
				f,
			),
		)
	}

	return prometheus.NewUntypedFunc(opts, untyped)
}
