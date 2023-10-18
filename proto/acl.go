package proto

import "github.com/btcsuite/btcutil/base58"

// AddrString returns the address string
func (x *Address) AddrString() string {
	return base58.CheckEncode(x.Address[1:], x.Address[0])
}

// Addr returns the address
func (x *AclResponse) Addr() (out [32]byte) {
	if x.Address == nil {
		return [32]byte{}
	}
	copy(out[:], x.Address.Address.Address[:32])
	return out
}
