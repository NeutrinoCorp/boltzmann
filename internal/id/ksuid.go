package id

import "github.com/segmentio/ksuid"

var _ FactoryFunc = NewKSUID

func NewKSUID() (string, error) {
	id, err := ksuid.NewRandom()
	return id.String(), err
}
