package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/atomyze-foundation/foundation/core/types"
	"github.com/atomyze-foundation/foundation/core/types/big"
	"github.com/atomyze-foundation/foundation/proto"
)

var (
	ErrBigIntFromString    = errors.New("big int from string")
	ErrPlatformAdminOnly   = errors.New("platform admin only")
	ErrEmptyLockID         = errors.New("empty lock id")
	ErrReason              = errors.New("empty reason")
	ErrLockNotExists       = errors.New("lock not exists")
	ErrAddressRequired     = errors.New("address required")
	ErrAmountRequired      = errors.New("amount required")
	ErrTokenTickerRequired = errors.New("token ticker required")
	ErrAlredyExist         = errors.New("lock alredy exist")
	ErrInsufficientFunds   = errors.New("insufficient funds to process")
)

const (
	BalanceTokenLockedEvent     = "BalanceTokenLocked"
	BalanceTokenUnlockedEvent   = "BalanceTokenUnlocked"
	BalanceAllowedLockedEvent   = "BalanceAllowedLocked"
	BalanceAllowedUnlockedEvent = "BalanceAllowedUnlocked"
)

// TxLockTokenBalance - блокирует токены на токенбалансе пользователя
// метод вызывает админ чейнкода, на вход приходит запрос BalanceLockRequest
func (bc *BaseContract) TxLockTokenBalance( //nolint:funlen
	sender *types.Sender,
	req *proto.BalanceLockRequest,
) error {
	if req.Id == "" {
		req.Id = bc.stub.GetTxID()
	}

	err := bc.verifyLockedArgs(sender, req)
	if err != nil {
		return err
	}

	// Проверить что уже есть
	_, err = bc.getLockedTokenBalance(req.Id)
	if err == nil {
		return ErrAlredyExist
	}

	address, err := types.AddrFromBase58Check(req.Address)
	if err != nil {
		return fmt.Errorf("address: %w", err)
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	if err = bc.TokenBalanceLock(address, amount); err != nil {
		return err
	}

	// state record with balance lock details
	balanceLock := &proto.TokenBalanceLock{
		Id:            req.Id,
		Address:       req.Address,
		Token:         req.Token,
		InitAmount:    req.Amount,
		CurrentAmount: req.Amount,
		Reason:        req.Reason,
		Docs:          req.Docs,
		Payload:       req.Payload,
	}

	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedToken)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{balanceLock.Id})
	if err != nil {
		return fmt.Errorf("create key: %w", err)
	}

	data, err := json.Marshal(balanceLock)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	balanceLockedEvent := &proto.TokenBalanceLocked{
		Id:      balanceLock.Id,
		Address: balanceLock.Address,
		Token:   balanceLock.Token,
		Amount:  balanceLock.CurrentAmount,
		Reason:  balanceLock.Reason,
		Docs:    balanceLock.Docs,
		Payload: balanceLock.Payload,
	}
	event, err := json.Marshal(balanceLockedEvent)
	if err != nil {
		return err
	}

	if err = bc.stub.SetEvent(BalanceTokenLockedEvent, event); err != nil {
		return err
	}

	return bc.stub.PutState(key, data)
}

// TxUnlockTokenBalance - раблокирует (полностью или частично) токены на токенбалансе пользователя
// метод вызывает админ чейнкода, на вход приходит запрос BalanceLockRequest
func (bc *BaseContract) TxUnlockTokenBalance( //nolint:funlen
	sender *types.Sender,
	req *proto.BalanceLockRequest,
) error {
	err := bc.verifyLockedArgs(sender, req)
	if err != nil {
		return err
	}

	// Проверить что уже есть
	balanceLock, err := bc.getLockedTokenBalance(req.Id)
	if err != nil {
		return err
	}

	address, err := types.AddrFromBase58Check(req.Address)
	if err != nil {
		return fmt.Errorf("address: %w", err)
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	cur, ok := new(big.Int).SetString(balanceLock.CurrentAmount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	isDelete := false
	c := cur.Cmp(amount)
	switch {
	case c < 0:
		return ErrInsufficientFunds
	case c == 0:
		isDelete = true
	}

	if err = bc.TokenBalanceUnlock(address, amount); err != nil {
		return err
	}

	// state record with balance lock details
	balanceLock.CurrentAmount = new(big.Int).Sub(cur, amount).String()

	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedToken)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{balanceLock.Id})
	if err != nil {
		return fmt.Errorf("create key: %w", err)
	}

	data, err := json.Marshal(balanceLock)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	balanceLockedEvent := &proto.TokenBalanceUnlocked{
		Id:                balanceLock.Id,
		Address:           balanceLock.Address,
		Token:             balanceLock.Token,
		Amount:            balanceLock.CurrentAmount,
		Reason:            balanceLock.Reason,
		Docs:              balanceLock.Docs,
		Payload:           balanceLock.Payload,
		CompleteOperation: isDelete,
	}
	event, err := json.Marshal(balanceLockedEvent)
	if err != nil {
		return err
	}

	if err = bc.stub.SetEvent(BalanceTokenUnlockedEvent, event); err != nil {
		return err
	}

	if isDelete {
		return bc.stub.DelState(key)
	}
	return bc.stub.PutState(key, data)
}

// QueryGetLockedTokenBalance - возвращает существующую блокировку токен баланса TokenBalanceLock
func (bc *BaseContract) QueryGetLockedTokenBalance(
	lockID string,
) (*proto.TokenBalanceLock, error) {
	return bc.getLockedTokenBalance(lockID)
}

// TxLockAllowedBalance - блокирует токены на аловедбалансе пользователя
// метод вызывает админ чейнкода, на вход приходит запрос BalanceLockRequest
func (bc *BaseContract) TxLockAllowedBalance( //nolint:funlen
	sender *types.Sender,
	req *proto.BalanceLockRequest,
) error {
	if req.Id == "" {
		req.Id = bc.stub.GetTxID()
	}

	err := bc.verifyLockedArgs(sender, req)
	if err != nil {
		return err
	}

	// Проверить что уже есть
	_, err = bc.getLockedAllowedBalance(req.Id)
	if err == nil {
		return ErrAlredyExist
	}

	address, err := types.AddrFromBase58Check(req.Address)
	if err != nil {
		return fmt.Errorf("address: %w", err)
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	if err = bc.AllowedBalanceLock(req.Token, address, amount); err != nil {
		return err
	}

	// state record with balance lock details
	balanceLock := &proto.AllowedBalanceLock{
		Id:            req.Id,
		Address:       req.Address,
		Token:         req.Token,
		InitAmount:    req.Amount,
		CurrentAmount: req.Amount,
		Reason:        req.Reason,
		Docs:          req.Docs,
		Payload:       req.Payload,
	}

	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedAllowed)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{balanceLock.Id})
	if err != nil {
		return fmt.Errorf("create key: %w", err)
	}

	data, err := json.Marshal(balanceLock)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	balanceLockedEvent := &proto.AllowedBalanceLocked{
		Id:      balanceLock.Id,
		Address: balanceLock.Address,
		Token:   balanceLock.Token,
		Amount:  balanceLock.CurrentAmount,
		Reason:  balanceLock.Reason,
		Docs:    balanceLock.Docs,
		Payload: balanceLock.Payload,
	}
	event, err := json.Marshal(balanceLockedEvent)
	if err != nil {
		return err
	}

	if err = bc.stub.SetEvent(BalanceAllowedLockedEvent, event); err != nil {
		return err
	}

	return bc.stub.PutState(key, data)
}

// TxUnlockAllowedBalance - раблокирует (полностью или частично) токены на аловедбалансе пользователя
// метод вызывает админ чейнкода, на вход приходит запрос BalanceLockRequest
func (bc *BaseContract) TxUnlockAllowedBalance( //nolint:funlen
	sender *types.Sender,
	req *proto.BalanceLockRequest,
) error {
	err := bc.verifyLockedArgs(sender, req)
	if err != nil {
		return err
	}

	// Проверить что уже есть
	balanceLock, err := bc.getLockedAllowedBalance(req.Id)
	if err != nil {
		return err
	}

	address, err := types.AddrFromBase58Check(req.Address)
	if err != nil {
		return fmt.Errorf("address: %w", err)
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	cur, ok := new(big.Int).SetString(balanceLock.CurrentAmount, 10) //nolint:gomnd
	if !ok {
		return ErrBigIntFromString
	}

	isDelete := false
	c := cur.Cmp(amount)
	switch {
	case c < 0:
		return ErrInsufficientFunds
	case c == 0:
		isDelete = true
	}

	if err = bc.AllowedBalanceUnLock(balanceLock.Token, address, amount); err != nil {
		return err
	}

	// state record with balance lock details
	balanceLock.CurrentAmount = new(big.Int).Sub(cur, amount).String()

	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedAllowed)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{balanceLock.Id})
	if err != nil {
		return fmt.Errorf("create key: %w", err)
	}

	data, err := json.Marshal(balanceLock)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	balanceLockedEvent := &proto.AllowedBalanceUnlocked{
		Id:                balanceLock.Id,
		Address:           balanceLock.Address,
		Token:             balanceLock.Token,
		Amount:            balanceLock.CurrentAmount,
		Reason:            balanceLock.Reason,
		Docs:              balanceLock.Docs,
		Payload:           balanceLock.Payload,
		CompleteOperation: isDelete,
	}
	event, err := json.Marshal(balanceLockedEvent)
	if err != nil {
		return err
	}

	if err = bc.stub.SetEvent(BalanceAllowedUnlockedEvent, event); err != nil {
		return err
	}

	if isDelete {
		return bc.stub.DelState(key)
	}
	return bc.stub.PutState(key, data)
}

// QueryGetLockedAllowedBalance - возвращает существующую блокировку аловед баланса AllowedBalanceLock
func (bc *BaseContract) QueryGetLockedAllowedBalance(
	lockID string,
) (*proto.AllowedBalanceLock, error) {
	return bc.getLockedAllowedBalance(lockID)
}

func (bc *BaseContract) getLockedTokenBalance(lockID string) (*proto.TokenBalanceLock, error) {
	if lockID == "" {
		return nil, ErrEmptyLockID
	}
	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedToken)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{lockID})
	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	data, err := bc.stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("get token balance lock from state: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("lock id=%s: %w", lockID, ErrLockNotExists)
	}

	balanceLock := &proto.TokenBalanceLock{}
	if err = json.Unmarshal(data, balanceLock); err != nil {
		return nil, fmt.Errorf("unmarshal token balance lock state: %w", err)
	}

	return balanceLock, nil
}

func (bc *BaseContract) getLockedAllowedBalance(lockID string) (*proto.AllowedBalanceLock, error) {
	if lockID == "" {
		return nil, ErrEmptyLockID
	}
	prefix := hex.EncodeToString([]byte{byte(StateKeyExternalLockedAllowed)})
	key, err := bc.stub.CreateCompositeKey(prefix, []string{lockID})
	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}

	data, err := bc.stub.GetState(key)
	if err != nil {
		return nil, fmt.Errorf("get allowed balance lock from state: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("lock id=%s: %w", lockID, ErrLockNotExists)
	}

	balanceLock := &proto.AllowedBalanceLock{}
	if err = json.Unmarshal(data, balanceLock); err != nil {
		return nil, fmt.Errorf("unmarshal allowed balance lock state: %w", err)
	}

	return balanceLock, nil
}

func (bc *BaseContract) verifyLockedArgs(
	sender *types.Sender,
	req *proto.BalanceLockRequest,
) error {
	// Проверка отправителя
	l := bc.GetInitArgsLen()
	if l < 1 {
		return ErrPlatformAdminOnly
	}

	admin, err := types.AddrFromBase58Check(bc.GetInitArg(argPositionAdmin))
	if err != nil {
		return err
	}

	if !sender.Equal(admin) {
		return ErrPlatformAdminOnly
	}

	// Проверка запроса
	if req.Id == "" {
		return ErrEmptyLockID
	}

	if req.Address == "" {
		return ErrAddressRequired
	}

	if req.Amount == "" {
		return ErrAmountRequired
	}

	if req.Token == "" {
		return ErrTokenTickerRequired
	}

	if req.Reason == "" {
		return ErrReason
	}

	return nil
}
