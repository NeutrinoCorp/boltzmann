package boltzmann

import (
	"fmt"

	"github.com/modern-go/reflect2"
)

// Retryable wraps a parent error to indicate its execution process is able to get retried.
type Retryable struct {
	Parent error
}

var _ error = Retryable{}

func (r Retryable) Error() string {
	return r.Parent.Error()
}

// ErrItemNotFound indicates if an item lookup operation failed due a non-existent entity in the storage.
type ErrItemNotFound struct {
	ResourceName string
	ResourceKey  string
}

var _ error = ErrItemNotFound{}

func (e ErrItemNotFound) Error() string {
	return fmt.Sprintf("item <<%s>> with key <<%s>> was not found", e.ResourceName, e.ResourceKey)
}

// ErrItemAlreadyExists indicates if an item write operation failed due existent entity already present in the storage.
type ErrItemAlreadyExists struct {
	ResourceName string
	ResourceKey  string
}

var _ error = ErrItemAlreadyExists{}

func (e ErrItemAlreadyExists) Error() string {
	return fmt.Sprintf("item <<%s>> with key <<%s>> already exists", e.ResourceName, e.ResourceKey)
}

// ErrOutOfRange indicates if a property is out of a specified range.
type ErrOutOfRange[T any] struct {
	PropertyName string
	A, B         T
}

var _ error = ErrOutOfRange[string]{}

// NewOutOfRangeWithType allocates a new ErrOutOfRange using T as base structure of a property.
// This routine will generate an ErrOutOfRange.PropertyName as `T_string_type.property`.
func NewOutOfRangeWithType[T, P any](property string, a, b P) ErrOutOfRange[P] {
	var zeroVal T
	typeOfStr := reflect2.TypeOf(zeroVal).String()
	return ErrOutOfRange[P]{
		PropertyName: fmt.Sprintf("%s.%s", typeOfStr, property),
		A:            a,
		B:            b,
	}
}

func (e ErrOutOfRange[T]) Error() string {
	return fmt.Sprintf("property <<%s>> has an invalid range (%v,%v)", e.PropertyName, e.A, e.B)
}
