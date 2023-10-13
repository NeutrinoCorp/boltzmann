package agent

import "fmt"

type ErrDriverNotFound struct {
	Driver string
}

var _ error = ErrDriverNotFound{}

func (e ErrDriverNotFound) Error() string {
	return fmt.Sprintf("agent: driver %s not found", e.Driver)
}
