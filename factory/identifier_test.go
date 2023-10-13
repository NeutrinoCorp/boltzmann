package factory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/neutrinocorp/boltzmann/factory"
)

func TestKSUID_NewID(t *testing.T) {
	out, err := factory.KSUID{}.NewID()
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
}
