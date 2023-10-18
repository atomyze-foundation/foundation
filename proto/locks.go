package proto

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// ConvertToCall - converts string into emission request, see foundation/core/reflect.go
func (x *BalanceLockRequest) ConvertToCall(
	_ shim.ChaincodeStubInterface,
	in string,
) (*BalanceLockRequest, error) {
	err := json.Unmarshal([]byte(in), x)
	return x, err
}
