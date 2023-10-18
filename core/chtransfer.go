package core

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/atomyze-foundation/foundation/core/cctransfer"
	"github.com/atomyze-foundation/foundation/core/types"
	"github.com/atomyze-foundation/foundation/core/types/big"
	pb "github.com/atomyze-foundation/foundation/proto"
)

const (
	argPositionAdmin = 0 // обычно для всех токено на этом месте ключ Админа
)

type typeOperation int

const (
	CreateFrom typeOperation = iota
	CreateTo
	CancelFrom
)

// TxChannelTransferByCustomer - транзакция инициирущая трансфер между каналами.
// Подписывает владелец токенов. Переводятся токены самому себе.
// После проверок, создается запись о трансфере и уменьшаются балансы у пользователя.
func (bc *BaseContract) TxChannelTransferByCustomer(
	sender *types.Sender,
	idTransfer string,
	to string,
	token string,
	amount *big.Int,
) (string, error) {
	return bc.createCCTransferFrom(idTransfer, to, sender.Address(), token, amount)
}

// TxChannelTransferByAdmin - транзакция инициирущая трансфер между каналами.
// Подписывает админ канала (площадка). Переводятся токены от пользователя idUser ему же.
// После проверок, создается запись о трансфере и уменьшаются балансы у пользователя.
func (bc *BaseContract) TxChannelTransferByAdmin(
	sender *types.Sender,
	idTransfer string,
	to string,
	idUser *types.Address,
	token string,
	amount *big.Int,
) (string, error) {
	// Проверки
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

	// Выполнение
	stub := bc.GetStub()

	// проверим, вдруг уже есть
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

	// изменение балансов
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

// TxCreateCCTransferTo - транзакция создает трансфер (уже с признаком коммит) в канале To
// и увеличивает балансы пользователю.
// Транзакция должна быть исполнена после инициирующей транзакции трансфера
// (TxChannelTransferByAdmin или TxChannelTransferByCustomer).
// Данная транзакция отправляется только сервисом channel-transfer с сертификатом "робота"
func (bc *BaseContract) TxCreateCCTransferTo(dataIn string) (string, error) {
	var tr pb.CCTransfer
	if err := proto.Unmarshal([]byte(dataIn), &tr); err != nil {
		if err = json.Unmarshal([]byte(dataIn), &tr); err != nil {
			return "", err
		}
	}

	// проверим, вдруг уже есть
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

	// изменение балансов
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

// TxCancelCCTransferFrom - транзакция отменяет (удаляет) запись о трансфере в канале From
// возвращает балансы пользователю. Если в течении некоторого таймаута сервис не может создать
// ответную часть в канале To, то требуется отменить трансфер.
// После TxChannelTransferByAdmin или TxChannelTransferByCustomer
// Данная транзакция отправляется только сервисом channel-transfer с сертификатом "робота"
func (bc *BaseContract) TxCancelCCTransferFrom(id string) error {
	// проверим, вдруг уже нет
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// если уже закомичено, то ошибка
	if tr.IsCommit {
		return cctransfer.ErrTransferCommit
	}

	// изменение балансов
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

// NBTxCommitCCTransferFrom - транзакция записывает флаг коммита в трансфере в канале From.
// Выполняется после успешного создания ответной части в канале To (TxCreateCCTransferTo)
// Данная транзакция отправляется только сервисом channel-transfer с сертификатом "робота"
func (bc *BaseContract) NBTxCommitCCTransferFrom(id string) error {
	// проверим, вдруг уже нет
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// если уже закомичено, то ошибка
	if tr.IsCommit {
		return cctransfer.ErrTransferCommit
	}

	tr.IsCommit = true
	return cctransfer.SaveCCFromTransfer(bc.GetStub(), tr)
}

// NBTxDeleteCCTransferFrom - транзакция удаляет запись о трансфере в канале From.
// Выполняется после успешного удаления в канале To (NBTxDeleteCCTransferTo)
// Данная транзакция отправляется только сервисом channel-transfer с сертификатом "робота"
func (bc *BaseContract) NBTxDeleteCCTransferFrom(id string) error {
	// проверим, вдруг уже нет
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// если не закомичено, то ошибка
	if !tr.IsCommit {
		return cctransfer.ErrTransferNotCommit
	}

	return cctransfer.DelCCFromTransfer(bc.GetStub(), id)
}

// NBTxDeleteCCTransferTo - транзакция удаляет запись о трансфере в канале To.
// Выполняется после успешного коммита в канале From (NBTxCommitCCTransferFrom)
// Данная транзакция отправляется только сервисом channel-transfer с сертификатом "робота"
func (bc *BaseContract) NBTxDeleteCCTransferTo(id string) error {
	// проверим, вдруг уже нет
	tr, err := cctransfer.LoadCCToTransfer(bc.GetStub(), id)
	if err != nil {
		return cctransfer.ErrNotFound
	}

	// если не закомичено, то ошибка
	if !tr.IsCommit {
		return cctransfer.ErrTransferNotCommit
	}

	return cctransfer.DelCCToTransfer(bc.GetStub(), id)
}

// QueryChannelTransferFrom - получие записи о трансфере из канала From
func (bc *BaseContract) QueryChannelTransferFrom(id string) (*pb.CCTransfer, error) {
	tr, err := cctransfer.LoadCCFromTransfer(bc.GetStub(), id)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// QueryChannelTransferTo - получие записи о трансфере из канала To
func (bc *BaseContract) QueryChannelTransferTo(id string) (*pb.CCTransfer, error) {
	tr, err := cctransfer.LoadCCToTransfer(bc.GetStub(), id)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// QueryChannelTransfersFrom - получие всех записей о трансферах из канала From
// Можно получать их частями (чанками)
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

	// ForwardDirection (Направление трансфера) - дополнительная переменная, сделанная для удобства
	// чтобы каждый раз её не вычислять. Вычисляется она 1 раз при заполнении структуры
	// при выполнении транзакции.
	// В зависимости от направления изменяются различные балансы.
	// Примеры:
	// Прямой трансфер: из канала A в канал B переводим токены A
	// или из канала B в канал A переводим токены B
	// Обратный трансфер:из канала A в канал B переводим токены B
	// или из канала B в канал A переводим токены A
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
