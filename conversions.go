// SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package touchstone

import (
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// collectorType is the cached reflect.Type for a prometheus.Collector
	collectorType = reflect.TypeOf((*prometheus.Collector)(nil)).Elem()
)

// CollectorAs attempts to set target to the given collector.  The target parameter
// must be a non-nil pointer to an interface or a prometheus metric type that implements
// Collector.  This function returns true if target was set, false to indicate no conversion
// was possitble.
//
// As it was inspired by errors.As, this function panics in similar situations where
// errors.As would panic:
//
//   - if target is the nil interface, e.g. CollectorAs(myMetric, nil) will panic
//   - if target is not both a pointer and non-nil
//   - if *target is not either:
//   - a concrete type that implements prometheus.Collector (e.g. prometheus.CounterVec), OR
//   - an arbitrary interface
func CollectorAs(c prometheus.Collector, target interface{}) bool {
	if target == nil {
		panic("touchstone.CollectorAs: target must not be a nil interface")
	}

	tValue := reflect.ValueOf(target)
	if tValue.Kind() != reflect.Ptr || tValue.IsNil() {
		panic("touchstone.CollectorAs: target must be a non-nil pointer")
	}

	tElem := tValue.Elem()
	if tElem.Kind() != reflect.Interface && !tElem.Type().Implements(collectorType) {
		panic("touchstone.CollectorAs: *target must either (1) be an interface or (2) implement prometheus.Collector")
	}

	cvalue := reflect.ValueOf(c)
	assignable := cvalue.Type().AssignableTo(tElem.Type())
	if assignable {
		tElem.Set(cvalue)
	}

	return assignable
}
