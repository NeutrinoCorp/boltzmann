package factory

import "github.com/segmentio/ksuid"

type Identifier interface {
	NewID() (string, error)
}

type KSUID struct{}

var _ Identifier = KSUID{}

func (i KSUID) NewID() (string, error) {
	return ksuid.New().String(), nil
}
