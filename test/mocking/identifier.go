package mocking

import (
	"github.com/stretchr/testify/mock"

	"github.com/neutrinocorp/boltzmann/factory"
)

type FakeIDFactory struct {
	mock.Mock
}

var _ factory.Identifier = &FakeIDFactory{}

func (f *FakeIDFactory) NewID() (string, error) {
	args := f.Called()
	return args.Get(0).(string), args.Error(1)
}
