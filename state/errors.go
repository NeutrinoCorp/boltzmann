package state

import "errors"

var (
	ErrTaskStateNotFound = errors.New("state.storage: task state not found")
)
