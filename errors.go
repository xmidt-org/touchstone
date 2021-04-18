package touchstone

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

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
