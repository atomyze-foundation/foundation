package core

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shimtest" //nolint:staticcheck
	"github.com/stretchr/testify/assert"
)

func TestQueryStub(t *testing.T) {
	stub := shimtest.NewMockStub("query", nil)

	stub.MockTransactionStart("txID")
	err := stub.PutState("key", []byte("value"))
	assert.NoError(t, err)
	stub.MockTransactionEnd("txID")

	queryStub := newQueryStub(stub)

	t.Run("GetState [positive]", func(t *testing.T) {
		val, _ := queryStub.GetState("key")
		assert.Equal(t, "value", string(val))
	})

	t.Run("PutState [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.PutState("key", []byte(""))
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("DelState [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.DelState("key")
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("SetStateValidationParameter [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.SetStateValidationParameter("key", []byte("new"))
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	err = stub.PutPrivateData("collection", "key", []byte("value"))
	assert.NoError(t, err)

	t.Run("PutPrivateData [negative]", func(t *testing.T) {
		t.Skip()
		err = queryStub.PutPrivateData("collection", "key", []byte("value2"))
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("DelPrivateData [negative]", func(t *testing.T) {
		t.Skip()
		err = queryStub.DelPrivateData("collection", "key")
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("PurgePrivateData [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.PurgePrivateData("collection", "key")
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("SetPrivateDataValidationParameter [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.SetPrivateDataValidationParameter("collection", "key", []byte("new"))
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})

	t.Run("SetEvent [negative]", func(t *testing.T) {
		t.Skip()
		err := queryStub.SetEvent("event", []byte("payload"))
		assert.Errorf(t, err, ErrMethodNotImplemented)
	})
}
