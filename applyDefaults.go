package touchstone

import "reflect"

// ApplyDefaults ensures that any field in dst that is the zero value takes
// a default from the corresponding field in src.  If a field in dst doesn't
// exist in src, that field is skipped.
//
// The src parameter may be the nil interface or a nil pointer to a struct, in
// which case no defaults are applied to dst.  Otherwise, src must be a struct
// or a pointer to a struct.
//
// The dst parameter must be a non-nil pointer to struct.  It cannot be the nil
// interface.
//
// Any other permutation of dst and src will result in a panic.
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
//   ApplyDefaults(&co, defaults)
//   c := prometheus.NewCounter(co)
//
// Note that this function does a shallow copy of any relevant fields.  In particular,
// that means that a slice of buckets will point to the same data in the dst and src
// after this function returns.
func ApplyDefaults(dst, src interface{}) {
	if src == nil {
		// allow nil defaults, which lets client code leave
		// defaults unset
		return
	} else if dst == nil {
		panic("touchstone.ApplyDefaults: dst must be a non-nil interface")
	}

	sValue := reflect.ValueOf(src)
	if sValue.Kind() == reflect.Ptr && sValue.Type().Elem().Kind() == reflect.Struct {
		if sValue.IsNil() {
			// allow nil pointer to struct, which lets unset pointers
			// to be the same as no defaults
			return
		}

		sValue = sValue.Elem() // dereference
	}

	if sValue.Kind() != reflect.Struct {
		panic("touchstone.ApplyDefaults: src must refer to a struct")
	}

	dValue := reflect.ValueOf(dst)
	if dValue.Kind() != reflect.Ptr || dValue.IsNil() || dValue.Type().Elem().Kind() != reflect.Struct {
		panic("touchstone.ApplyDefaults: dst must be a non-nil pointer to struct")
	}

	dValue = dValue.Elem() // dereference
	sType := sValue.Type()
	dType := dValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		sField := sType.Field(i)
		if len(sField.PkgPath) > 0 || sField.Anonymous {
			// skip unexported or anonymous fields in src
			continue
		}

		sFieldValue := sValue.Field(i)
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

		if dFieldValue := dValue.FieldByName(dField.Name); dFieldValue.IsZero() {
			// shallow copy the field from src -> dst if and only if
			// the dst field is the zero value for its type
			dFieldValue.Set(sFieldValue)
		}
	}
}
