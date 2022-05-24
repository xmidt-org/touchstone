package touchbundle

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldError(t *testing.T) {
	t.Run("NoCause", func(t *testing.T) {
		fe := &FieldError{
			Field: reflect.StructField{
				Name: "Metric",
				Type: reflect.TypeOf(123), // doesn't matter
			},
			Message: "message",
		}

		assert := assert.New(t)
		assert.Nil(fe.Unwrap())
		assert.Contains(fe.Error(), "Metric")
		assert.Contains(fe.Error(), "int")
		assert.Contains(fe.Error(), "message")
	})

	t.Run("WithCause", func(t *testing.T) {
		expected := errors.New("expected")

		fe := &FieldError{
			Field: reflect.StructField{
				Name: "Metric",
				Type: reflect.TypeOf(123), // doesn't matter
			},
			Cause:   expected,
			Message: expected.Error(),
		}

		assert := assert.New(t)
		assert.ErrorIs(fe, expected)
		assert.Contains(fe.Error(), "Metric")
		assert.Contains(fe.Error(), "int")
		assert.Contains(fe.Error(), "expected")
	})
}
