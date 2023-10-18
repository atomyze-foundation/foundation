package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testContract struct {
	BaseContract
}

func (*testContract) TxTestFunction() error {
	return nil
}

func (*testContract) GetID() string {
	return "TEST"
}

func TestDisabledFunctions(t *testing.T) {
	cc1, err := NewCC(&testContract{}, nil)
	assert.NoError(t, err)
	_, exists1 := cc1.methods["testFunction"]
	assert.True(t, exists1)

	cc2, err := NewCC(&testContract{}, &ContractOptions{
		DisabledFunctions: []string{"TxTestFunction"},
	})
	assert.NoError(t, err)
	_, exists2 := cc2.methods["testFunction"]
	assert.False(t, exists2)
}
