package controller

import (
	"reflect"

	"github.com/iancoleman/strcase"
)

// ViewData is a general-purposed container for views to wrap themselves in a data field.
//
// Useful in HTTP APIs for response standardization.
type ViewData struct {
	Data map[string]any `json:"data"`
}

// NewViewData allocates a new ViewData instance.
// Uses T's type name as the key of the single entry for ViewData.Data map.
func NewViewData[T any](data any) ViewData {
	var zeroVal T
	typeOfStr := strcase.ToSnake(reflect.TypeOf(zeroVal).Name())
	return ViewData{
		Data: map[string]any{
			typeOfStr: data,
		},
	}
}
