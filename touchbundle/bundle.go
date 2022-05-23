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

package touchbundle

import (
	"fmt"
	"reflect"

	"github.com/xmidt-org/touchstone"
	"go.uber.org/fx"
	"go.uber.org/multierr"
)

// Bundle represents a group of metrics.  A bundle must always be a non-nil pointer to struct.
type Bundle interface{}

// populate is the common function for filling out a bundle struct.  The supplied reflect.Value
// must be an addressable, settable struct.
func populate(factory *touchstone.Factory, bundle reflect.Value) (err error) {
	for i := 0; i < bundle.NumField(); i++ {
		f := metricField(bundle.Type().Field(i))
		if f.skip() {
			continue
		}

		opts, labelNames, fieldErr := f.newOpts()
		err = multierr.Append(err, fieldErr)
		if opts == nil || fieldErr != nil {
			continue
		}

		var metric interface{}
		if len(labelNames) > 0 {
			metric, fieldErr = factory.NewVec(opts, labelNames...)
		} else {
			metric, fieldErr = factory.New(opts)
		}

		err = multierr.Append(err, fieldErr)
		if fieldErr == nil {
			bundle.Field(i).Set(reflect.ValueOf(metric))
		}
	}

	return
}

// Populate fills out a bundle with metrics created by the given Factory.
func Populate(f *touchstone.Factory, b Bundle) error {
	bv := reflect.ValueOf(b)
	if bv.Kind() == reflect.Ptr && !bv.IsNil() {
		bv = bv.Elem()
	}

	if bv.Kind() != reflect.Struct || !bv.CanAddr() {
		return fmt.Errorf(
			"'%T' is not a valid bundle.  It must be a non-nil pointer to a struct.",
			b,
		)
	}

	return populate(f, bv)
}

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	factoryType = reflect.TypeOf((*touchstone.Factory)(nil))
)

// Provide emits a bundle as an uber/fx component.  The supplied prototype must be
// either a struct or a pointer to struct.  The returned component will be a new
// instance of the same type as the prototype.
//
// For example:
//
//     app := fx.New(
//         touchbundle.Provide(MyMetrics{}),
//         fx.Invoke(
//             func(m MyMetrics) {
//                 // m's metric fields will have been populated
//             },
//         ),
//     )
//
//     app := fx.New(
//         touchbundle.Provide((*MyStruct)(nil)),
//         fx.Invoke(
//             func(m *MyMetrics) {
//                 // m's metric fields will have been populated
//                 // m will point to a distinct, new instance of MyMetrics
//             },
//         ),
//     )
func Provide(prototype interface{}) fx.Option {
	var (
		componentType = reflect.TypeOf(prototype)
		structType    reflect.Type
	)

	switch {
	case componentType.Kind() == reflect.Struct:
		structType = componentType

	case componentType.Kind() == reflect.Ptr && componentType.Elem().Kind() == reflect.Struct:
		structType = componentType.Elem()

	default:
		return fx.Error(
			fmt.Errorf(
				"'%T' is not a valid bundle prototype.  It is not a struct or pointer to struct.",
				prototype,
			),
		)
	}

	ctor := reflect.MakeFunc(
		reflect.FuncOf(
			[]reflect.Type{
				factoryType,
			},
			[]reflect.Type{componentType, errorType},
			false,
		),
		func(in []reflect.Value) (out []reflect.Value) {
			out = make([]reflect.Value, 2)
			var (
				factory     = in[0].Interface().(*touchstone.Factory)
				errValue    = reflect.New(errorType)
				bundleValue = reflect.New(structType)
				err         = populate(factory, bundleValue.Elem())
			)

			if err != nil {
				errValue.Elem().Set(
					reflect.ValueOf(err),
				)
			}

			if componentType.Kind() == reflect.Ptr {
				out[0] = bundleValue
			} else {
				out[0] = bundleValue.Elem()
			}

			out[1] = errValue.Elem()
			return
		},
	)

	return fx.Provide(ctor.Interface())
}
