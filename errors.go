package touchstone

import (
	"errors"
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
//     - a concrete type that implements prometheus.Collector (e.g. prometheus.CounterVec), OR
//     - an arbitrary interface
func CollectorAs(c prometheus.Collector, target interface{}) bool {
	if target == nil {
		panic("touchstone: target must not be a nil interface")
	}

	tValue := reflect.ValueOf(target)
	if tValue.Kind() != reflect.Ptr || tValue.IsNil() {
		panic("touchstone: target must be a non-nil pointer")
	}

	tElem := tValue.Elem()
	if tElem.Kind() != reflect.Interface && !tElem.Type().Implements(collectorType) {
		panic("touchstone: *target must either (1) be an interface or (2) implement prometheus.Collector")
	}

	cvalue := reflect.ValueOf(c)
	assignable := cvalue.Type().AssignableTo(tElem.Type())
	if assignable {
		tElem.Set(cvalue)
	}

	return assignable
}

// AsAlreadyRegisteredError tests if err is a prometheus AlreadyRegisteredError.
// If it is, a non-nil error is returned.  If the error wasn't an AlreadyRegisteredError,
// this method returns nil.
func AsAlreadyRegisteredError(err error) *prometheus.AlreadyRegisteredError {
	// NOTE: prometheus doesn't follow golang recommendations.  This error type
	// is not a pointer receiver for Error().
	var are prometheus.AlreadyRegisteredError
	if errors.As(err, &are) {
		return &are
	}

	return nil
}

// ExistingCollector attempts to use the previously registered collector as target.
// This function attempts to coerce err to a prometheus.AlreadyRegisteredError,
// and then coerce the ExistingCollector field to the target.
//
// If err was not a prometheus.AlreadyRegisteredError or if the existing collector
// was not assignable to the target, this function returns the original error.
// Otherwise, this function returns nil.  Note that target is completely ignored if
// err is not a prometheus.AlreadyRegisteredError, in which case this function
// returns nil.
//
// A typical use of this method is to allow client code to ignore already registered
// errors and just take the previously registered metric:
//
//   // using a touchstone Factory:
//   var f *touchstone.Factory
//   m, err := f.NewCounterVec(/* ... */)
//   err = touchstone.ExistingCollector(&m, err) // note the &m to replace m with the existing metric
//
//   // using a prometheus Registerer
//   r := prometheus.NewPedanticRegistry()
//   cv := prometheus.NewCounterVec(/* ... */)
//   err := r.Register(cv)
//   err = touchstone.ExistingCollector(&cv, err) // note the &cv to replace cv with the existing counter vec
func ExistingCollector(target interface{}, err error) error {
	if are := AsAlreadyRegisteredError(err); are != nil {
		if CollectorAs(are.ExistingCollector, target) {
			return nil
		}
	}

	return err
}
