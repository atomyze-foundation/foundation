package core

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/atomyze-foundation/foundation/core/cctransfer"
	"github.com/atomyze-foundation/foundation/core/types"
	"github.com/atomyze-foundation/foundation/core/types/big"
	pb "github.com/atomyze-foundation/foundation/proto"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

const (
	argPositionAdmin = 0 // Usually for all tokens, this is where the Admin key is
)

type typeOperation int

const (
	// CreateFrom - Channel transference creation From
	CreateFrom typeOperation = iota
	// CreateTo - Channel transference creation To
	CreateTo
	// CancelFrom - cancellation in the From
	CancelFrom
)

// TxChannelTransferByCustomer - transaction initiating transfer between channels.
// The owner of tokens signs. Tokens are transferred to themselveselves.
// After the checks, a transfer record is created and the user's balances are reduced.
func (bc *BaseContract) TxChannelTransferByCustomer(
	sender *types.Sender,
	idTransfer string,
	to string,
	token string,
	amount *big.Int,
) (string, error) {
	return bc.createCCTransferFrom(idTransfer, to, sender.Address(), token, amount)
}

// TxChannelTransferByAdmin - transaction initiating transfer between channels.
// Signed by the channel admin (site). The tokens are transferred from idUser to the same user.
// After the checks, a transfer record is created and the user's balances are reduced.
func (bc *BaseContract) TxChannelTransferByAdmin(
	sender *types.Sender,
	idTransfer string,
	to string,
	idUser *types.Address,
	token string,
	amount *big.Int,
) (string, error) {
	// Checks
	l := bc.GetInitArgsLen()
	if l < 1 {
		return "", cctransfer.ErrNotFoundAdminKey
	}

	admin, err := types.AddrFromBase58Check(bc.GetInitArg(argPositionAdmin))
	if err != nil {
		panic(err)
	}

	if !sender.Equal(admin) {
		return "", cctransfer.ErrNotFoundAdminKey
	}

	if sender.Equal(idUser) {
		return "", cctransfer.ErrInvalidIDUser
	}

	return bc.createCCTransferFrom(idTransfer, to, idUser, token, amount)
}

func (bc *BaseContract) createCCTransferFrom(
	idTransfer string,
	to string,
	idUser *types.Address,
	token string,
	amount *big.Int,
) (string, error) {
	if strings.EqualFold(bc.id, to) {
		return "", cctransfer.ErrInvalidChannel
	}

	t := tokenSymbol(token)

	if !strings.EqualFold(bc.id, t) && !strings.EqualFold(to, t) {
		return "", cctransfer.ErrInvalidToken
	}

	// Fulfillment
	stub := bc.GetStub()

	// see if it's already there.
	if _, err := cctransfer.LoadCCFromTransfer(stub, idTransfer); err == nil {
		return "", cctransfer.ErrIDTransferExist
	}

	ts, err := stub.GetTxTimestamp()
	if err != nil {
		return "", err
	}

	tr := &pb.CCTransfer{
		Id:               idTransfer,
		From:             bc.id,
		To:               to,
		Token:            token,
		User:             idUser.Bytes(),
		Amount:           amount.Bytes(),
		ForwardDirection: strings.EqualFold(bc.id, t),
		TimeAsNanos:      ts.AsTime().UnixNano(),
	}

	if err = cctransfer.SaveCCFromTransfer(stub, tr); err != nil {
		return "", err
	}

	// rebalancing
	err = bc.ccTransferChangeBalance(
		CreateFrom,
		tr.ForwardDirection,
		idUser,
		amount,
		tr.From,
		tr.To,
		tr.Token,
	)
	if err != nil {
		return "", err
	}

	return bc.GetStub().GetTxID(), nil
}

// TxCreateCCTransferTo - transaction creates a transfer (already with commit sign) in the channel To
// and increases the user's balances.
// The transaction must be executed after the initiating transfer transaction
// (TxChannelTransferByAdmin or TxChannelTransferByCustomer).
// This transaction is sent only by the channel-transfer service with a "robot" certificate
func (bc *BaseContract) TxCreateCCTransferTo(dataIn string) (string, error) {
	var tr pb.CCTransfer
	if err := proto.Unmarshal([]byte(dataIn), &tr); err != nil {
		if err = json.Unmarshal([]byte(dataIn), &tr); err != nil {
			return "", err
		}
	}

	// see if it's already there.
	if _, err := cctransfer.LoadCCToTransfer(bc.GetStub(), tr.Id); err == nil {
		return "", cctransfer.ErrIDTransferExist
	}

	if !strings.EqualFold(bc.id, tr.From) && !strings.EqualFold(bc.id, tr.To) {
		return "", cctransfer.ErrInvalidChannel
	}

	if strings.EqualFold(tr.From, tr.To) {
		return "", cctransfer.ErrInvalidChannel
	}

	t := tokenSymbol(tr.Token)

	if !strings.EqualFold(tr.From, t) && !strings.EqualFold(tr.To, t) {
		return "", cctransfer.ErrInvalidToken
	}

	if strings.EqualFold(tr.From, t) != tr.ForwardDirection {
		return "", cctransfer.ErrInvalidToken
	}

	tr.IsCommit = true
	if err := cctransfer.SaveCCToTransfer(bc.GetStub(), &tr); err != nil {
		return "", err
	}

	// rebalancing
	err := bc.ccTransferChangeBalance(
		CreateTo,
		tr.ForwardDirection,
		types.AddrFromBytes(tr.User),
		new(big.Int).SetBytes(tr.Amount),
		tr.From,
		tr.To,
		tr.Token,
	)
	if err != nil {
		return "", err
	}

	return bc.GetStub().GetTxID(), nil
}

// TxCancelCCTransferFrom - transaction cancels (deletes) the transfer record in the From channel
// returns balances to the user. If the service cannot create a response part in the "To" channel
// within some timeout, it is required to cancel the transfer.
// After TxChannelTransferByAdmin or TxChannelTransferByCustomer
// This transaction is sent only by the channel-transfer service with a "robot" certificate
func (bc *BaseContract) TxCancelCCTransferFrom(id string) error {
	// see if it's already gone
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// If it's already committed, it's a mistake.
	if tr.IsCommit {
		return cctransfer.ErrTransferCommit
	}

	// rebalancing
	err = bc.ccTransferChangeBalance(
		CancelFrom,
		tr.ForwardDirection,
		types.AddrFromBytes(tr.User),
		new(big.Int).SetBytes(tr.Amount),
		tr.From,
		tr.To,
		tr.Token,
	)
	if err != nil {
		return err
	}

	return cctransfer.DelCCFromTransfer(bc.GetStub(), id)
}

// NBTxCommitCCTransferFrom - transaction writes the commit flag in the transfer in the From channel.
// Executed after successful creation of a mating part in the channel To (TxCreateCCTransferTo)
// This transaction is sent only by the channel-transfer service with a "robot" certificate
func (bc *BaseContract) NBTxCommitCCTransferFrom(id string) error {
	// see if it's already gone
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// if it's already committed, it's an error
	if tr.IsCommit {
		return cctransfer.ErrTransferCommit
	}

	tr.IsCommit = true
	return cctransfer.SaveCCFromTransfer(bc.GetStub(), tr)
}

// NBTxDeleteCCTransferFrom - transaction deletes the transfer record in the channel From.
// Performed after successful removal in the canal To (NBTxDeleteCCTransferTo)
// This transaction is sent only by the channel-transfer service with a "robot" certificate
func (bc *BaseContract) NBTxDeleteCCTransferFrom(id string) error {
	// see if it's already gone
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// if it's not committed, it's an error
	if !tr.IsCommit {
		return cctransfer.ErrTransferNotCommit
	}

	return cctransfer.DelCCFromTransfer(bc.GetStub(), id)
}

// NBTxDeleteCCTransferTo - transaction deletes transfer record in channel To.
// Executed after a successful commit in the From channel (NBTxCommitCCTransferFrom)
// This transaction is sent only by the channel-transfer service with a "robot" certificate
func (bc *BaseContract) NBTxDeleteCCTransferTo(id string) error {
	// Let's check if it's not already
	tr, err := cctransfer.LoadCCToTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// if it is not committed, error
	if !tr.IsCommit {
		return cctransfer.ErrTransferNotCommit
	}

	return cctransfer.DelCCToTransfer(bc.GetStub(), id)
}

// QueryChannelTransferFrom - receiving a transfer record from the channel From
func (bc *BaseContract) QueryChannelTransferFrom(id string) (*pb.CCTransfer, error) {
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// QueryChannelTransferTo - receiving a transfer record from the channel To
func (bc *BaseContract) QueryChannelTransferTo(id string) (*pb.CCTransfer, error) {
	tr, err := cctransfer.LoadCCToTransfer(bc.GetStub(), id)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// QueryChannelTransfersFrom - getting all transfer records from the channel From
// You can receive them in parts (chunks)
func (bc *BaseContract) QueryChannelTransfersFrom(pageSize int64, bookmark string) (*pb.CCTransfers, error) {
	if pageSize <= 0 {
		return nil, cctransfer.ErrPageSizeLessOrEqZero
	}

	prefix := cctransfer.CCFromTransfers()
	startKey, endKey := prefix, prefix+string(utf8.MaxRune)

	if bookmark != "" && !strings.HasPrefix(bookmark, prefix) {
		return nil, cctransfer.ErrInvalidBookmark
	}

	trs, err := cctransfer.LoadCCFromTransfers(bc.GetStub(), startKey, endKey, bookmark, int32(pageSize))
	if err != nil {
		return nil, err
	}

	return trs, nil
}

func (bc *BaseContract) ccTransferChangeBalance( //nolint:funlen,gocognit
	t typeOperation,
	forwardDirection bool,
	user *types.Address,
	amount *big.Int,
	from string,
	to string,
	token string,
) error {
	var err error

	// ForwardDirection (Transfer direction) - is an additional variable made for convenience
	// to avoid calculating it every time. It is calculated once when filling the structure
	// when executing a transaction.
	// Depending on the direction, different balances change.
	// Examples:
	// Direct transfer: from channel A to channel B we transfer tokens A
	// or from channel B to channel A transfer tokens B
	// Reverse transfer: from channel A to channel B transfer tokens B
	// or from channel B to channel A transfer tokens A
	switch t {
	case CreateFrom:
		if forwardDirection {
			if err = bc.tokenBalanceSub(user, amount, token); err != nil {
				return err
			}
			if err = GivenBalanceAdd(bc.GetStub(), to, amount); err != nil {
				return err
			}
		} else {
			if err = bc.AllowedBalanceSub(token, user, amount, "ch-transfer"); err != nil {
				return err
			}
		}
	case CreateTo:
		if forwardDirection {
			if err = bc.AllowedBalanceAdd(token, user, amount, "ch-transfer"); err != nil {
				return err
			}
		} else {
			if err = bc.tokenBalanceAdd(user, amount, token); err != nil {
				return err
			}
			if err = GivenBalanceSub(bc.GetStub(), from, amount); err != nil {
				return err
			}
		}
	case CancelFrom:
		if forwardDirection {
			if err = bc.tokenBalanceAdd(user, amount, token); err != nil {
				return err
			}
			if err = GivenBalanceSub(bc.GetStub(), to, amount); err != nil {
				return err
			}
		} else {
			if err = bc.AllowedBalanceAdd(token, user, amount, "cancel ch-transfer"); err != nil {
				return err
			}
		}
	default:
		return cctransfer.ErrUnauthorizedOperation
	}

	return nil
}

func tokenSymbol(token string) string {
	parts := strings.Split(token, "_")
	return parts[0]
}
