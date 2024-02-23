package boltzmann

import "fmt"

type ErrItemNotFound struct {
	ResourceName string
	ResourceID   string
}

var _ error = ErrItemNotFound{}

func (e ErrItemNotFound) Error() string {
	return fmt.Sprintf("Item <<%s>> with id <<%s>> was not found", e.ResourceName, e.ResourceID)
}
