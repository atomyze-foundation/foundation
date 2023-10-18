package token

import (
	"errors"

	"github.com/atomyze-foundation/foundation/core/types"
	"github.com/atomyze-foundation/foundation/core/types/big"
)

type VT struct {
	BaseToken
}

func (vt *VT) TxEmitToken(sender *types.Sender, amount *big.Int) error {
	if !sender.Equal(vt.Issuer()) {
		return errors.New("unauthorized")
	}
	if err := vt.TokenBalanceAdd(vt.Issuer(), amount, "emitToken"); err != nil {
		return err
	}
	return vt.EmissionAdd(amount)
}
