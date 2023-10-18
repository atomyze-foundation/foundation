package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/atomyze-foundation/foundation/core/helpers"
	"github.com/atomyze-foundation/foundation/core/types/big"
	pb "github.com/atomyze-foundation/foundation/proto"
)

// Address might be more complicated structure
// contains fields like isIndustrial bool or isMultisig bool

const addressLength = 32

type Address pb.Address

func AddrFromBytes(in []byte) *Address {
	addr := &Address{}
	addrBytes := make([]byte, addressLength)
	copy(addrBytes, in[:32])
	addr.Address = addrBytes
	return addr
}

func AddrFromBase58Check(in string) (*Address, error) {
	value, ver, err := base58.CheckDecode(in)
	if err != nil {
		return &Address{}, err
	}
	addr := &Address{}
	addrBytes := make([]byte, addressLength)
	copy(addrBytes, append([]byte{ver}, value...)[:32])
	addr.Address = addrBytes
	return addr, nil
}

func (a *Address) Equal(b *Address) bool {
	return bytes.Equal(a.Address, b.Address)
}

func (a *Address) Bytes() []byte {
	return a.Address
}

func (a *Address) String() string {
	return base58.CheckEncode(a.Address[1:], a.Address[0])
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Address) PrepareToSave(stub shim.ChaincodeStubInterface, in string) (string, error) {
	accInfo, err := helpers.GetAccountInfo(stub, in)
	if err != nil {
		return "", err
	}
	if accInfo.BlackListed {
		return "", fmt.Errorf("address %s is blacklisted", in)
	}
	return in, nil
}

func (a *Address) ConvertToCall(_ shim.ChaincodeStubInterface, in string) (*Address, error) { // stub
	// only this called in batch
	return AddrFromBase58Check(in)
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var tmp string
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	a1, err := AddrFromBase58Check(tmp)
	a.UserID = a1.UserID
	a.Address = a1.Address
	a.IsIndustrial = a1.IsIndustrial
	a.IsMultisig = a1.IsMultisig
	return err
}

func (a *Address) IsUserIDSame(b *Address) bool {
	if a.UserID == "" || b.UserID == "" {
		return false
	}
	return a.UserID == b.UserID
}

type Sender struct {
	addr *Address
}

func NewSenderFromAddr(addr *Address) *Sender {
	return &Sender{addr: addr}
}

func (s *Sender) Address() *Address {
	return s.addr
}

func (s *Sender) Equal(addr *Address) bool {
	return bytes.Equal(s.addr.Address, addr.Address)
}

type Hex []byte

func (h Hex) ConvertToCall(_ shim.ChaincodeStubInterface, in string) (Hex, error) { // stub
	value, err := hex.DecodeString(in)
	return value, err
}

type MultiSwapAssets struct {
	Assets []*MultiSwapAsset
}

type MultiSwapAsset struct {
	Group  string `json:"group,omitempty"`
	Amount string `json:"amount,omitempty"`
}

func ConvertToAsset(in []*MultiSwapAsset) ([]*pb.Asset, error) {
	if in == nil {
		return nil, errors.New("assets can't be nil")
	}

	assets := make([]*pb.Asset, 0, len(in))
	for _, item := range in {
		value, ok := new(big.Int).SetString(item.Amount, 10) //nolint:gomnd
		if !ok {
			return nil, fmt.Errorf("couldn't convert %s to bigint", item.Amount)
		}
		if value.Cmp(big.NewInt(0)) < 0 {
			return nil, fmt.Errorf("value %s should be positive", item.Amount)
		}

		asset := pb.Asset{}
		asset.Amount = value.Bytes()
		asset.Group = item.Group
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (n MultiSwapAssets) ConvertToCall(_ shim.ChaincodeStubInterface, in string) (MultiSwapAssets, error) { // stub
	assets := MultiSwapAssets{}
	err := json.Unmarshal([]byte(in), &assets)
	if err != nil {
		return assets, err
	}
	return assets, nil
}

func IsValidAddressLen(val []byte) bool {
	return len(val) == addressLength
}
