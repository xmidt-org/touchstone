package touchstone

import "reflect"

// copyDefaults copies any non-zero field in src to a zero field in dst.
// Unexported and anonymous fields are skipped.  Both dst and src
// must be structs, or this function panics.
func copyDefaults(dst, src reflect.Value) {
	sType := src.Type()
	dType := dst.Type()
	for i := 0; i < sType.NumField(); i++ {
		sField := sType.Field(i)
		if len(sField.PkgPath) > 0 || sField.Anonymous {
			// skip unexported or anonymous fields in src
			continue
		}

		sFieldValue := src.Field(i)
		if sFieldValue.IsZero() {
			// skip any src field that wasn't set
			continue
		}

		dField, present := dType.FieldByName(sField.Name)
		if !present || len(dField.PkgPath) > 0 || dField.Anonymous {
			// skip unexported or anonymous fields in dst
			// also skip any fields in src that are not present in dst
			//
			// NOTE: should never get unexported fields from dst at this
			// point, since we checked the field on src and we're using
			// the same name.  However, it's entirely possible to get
			// an anonymous field in dst with the same name as the src field.
			continue
		}

		if sField.Type != dField.Type {
			// skip when src and dst fields aren't of the same type
			continue
		}

		if dFieldValue := dst.FieldByName(dField.Name); dFieldValue.IsZero() {
			// shallow copy the field from src -> dst if and only if
			// the dst field is the zero value for its type
			dFieldValue.Set(sFieldValue)
		}
	}
}

// ApplyDefaults ensures that any field in dst that is the zero value takes
// a default from the corresponding field in src.  If a field in dst doesn't
// exist in src, that field is skipped.
//
// This function always returns a non-nil pointer to the same type of struct
// that dst refers to.  If dst is a struct value, this function creates a new
// instance of that struct, applies the defaults to the new instance, and returns
// a pointer to that new instance.  If dst is a non-nil pointer to struct, defaults
// are applied in-place to that struct and dst is returned as is.  If dst is any
// other type, including a nil pointer, this function panics.
//
// The src parameter may be the nil interface or a nil pointer to a struct, in
// which case no defaults are applied to dst.  However, the rules for dst still
// apply: if dst is a struct value a pointer to a new struct of that type is returned.
// Otherwise, src must be a struct or a pointer to a struct.  Any other type for src
// will result in a panic.
//
// The primary use case for this function is setting up default options for
// prometheus metrics:
//
//   // note that this can be any struct with fields named the same
//   // as the prometheus xxxOpts struct
//   defaults := prometheus.CounterOpts{
//     Namespace: "default",
//     // can set any other fields as defaults
//   }
//
//   co := prometheus.CounterOpts{
//     Name: "my_counter",
//   }
//
//   ApplyDefaults(&co, defaults) // in-place transfer to co
//   c := prometheus.NewCounter(co)
//
// The result of ApplyDefaults is safe for casting to *dst, even if dst is nil:
//
//   defaults := prometheus.Opts{
//     Namespace: "default",
//     Subsystem: "default",
//   }
//
//   // creates a new opts
//   co := ApplyDefaults((*prometheus.CounterOpts)(nil), defaults).(*prometheus.CounterOpts)
//
//   // creates a new opts which is a clone of dst
//   go := ApplyDefaults(prometheus.GaugeOpts{Name: "cloneme"}, defaults).(*prometheus.GaugeOpts)
//
// Note that this function does a shallow copy of any relevant fields.  In particular,
// that means that a slice of buckets will point to the same data in the dst and src
// after this function returns.
func ApplyDefaults(dst, src interface{}) (result interface{}) {
	sValue := reflect.ValueOf(src)
	if sValue.Kind() == reflect.Ptr {
		if sValue.Type().Elem().Kind() != reflect.Struct {
			panic("touchstone.ApplyDefaults: src must be nil, a pointer to struct, or a struct")
		}

		if !sValue.IsNil() {
			sValue = sValue.Elem() // dereference
		}
	} else if sValue.IsValid() && sValue.Kind() != reflect.Struct {
		panic("touchstone.ApplyDefaults: src must be nil, a pointer to struct, or a struct")
	}

	dValue := reflect.ValueOf(dst)
	if dValue.Kind() == reflect.Struct {
		// create a new struct and return a pointer to it
		pv := reflect.New(dValue.Type())
		pv.Elem().Set(dValue)
		dValue = pv.Elem() // dereference
		result = pv.Interface()
	} else if dValue.Kind() == reflect.Ptr && dValue.Type().Elem().Kind() == reflect.Struct {
		if dValue.IsNil() {
			// create a new struct and return a pointer to it
			// since the pointer was nil, we can't set the new struct's fields
			pv := reflect.New(dValue.Type().Elem())
			dValue = pv.Elem() // dereference
			result = pv.Interface()
		} else {
			// use the existing struct in place, and return that pointer
			result = dValue.Interface()
			dValue = dValue.Elem()
		}
	} else {
		panic("touchstone.ApplyDefaults: dst must be a struct or a pointer to struct")
	}

	if !sValue.IsValid() || (sValue.Kind() == reflect.Ptr && sValue.IsNil()) {
		// this covers both the case (1) where src == nil interface{}
		// and (2) where src is a nil pointer
		// we want to return the pointer to the dst value in this case,
		// so that client code is simpler
		return
	}

	//
	// now we can apply any applicable fields in src to the dst
	//

	copyDefaults(dValue, sValue)
	return
}
