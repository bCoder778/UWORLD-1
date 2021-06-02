package types

import (
	"github.com/uworldao/UWORLD/common/hasharry"
)

// Account Status
type IAccount interface {
	GetBalance(string) uint64
	GetNonce() uint64
	Update(uint64) error
	StateKey() hasharry.Address
	IsExist() bool
	IsNeedUpdate() bool
	TransferChangeFrom(ITransaction, uint64) error
	TransferV2ChangeFrom(ITransaction, uint64) error
	ContractChangeFrom(ITransaction, uint64) error
	TransferChangeTo(*Receiver, uint64, hasharry.Address, uint64) error
	TransferV2ChangeTo(*Receiver, hasharry.Address, uint64) error
	ContractChangeTo(*Receiver, hasharry.Address, uint64) error
	FeesChange(uint64, uint64)
	ConsumptionChange(uint64, uint64)
	TransferOut(token hasharry.Address, amount, height uint64) error
	TransferIn(token hasharry.Address, amount, height uint64) error
	VerifyTxState(ITransaction) error
	VerifyNonce(uint64) error
	IsEmpty() bool
}

type IChainAddress interface {
	AddressList() []string
}
