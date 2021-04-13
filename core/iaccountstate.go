package core

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
)

type IAccountState interface {
	InitTrie(stateRoot hasharry.Hash) error

	GetAccountState(stateKey hasharry.Address) types.IAccount

	GetAccountNonce(stateKey hasharry.Address) (uint64, error)

	UpdateTransferFrom(tx types.ITransaction, blockHeight uint64) error

	UpdateTransferV2From(tx types.ITransaction, blockHeight uint64) error

	UpdateContractFrom(tx types.ITransaction, blockHeight uint64) error

	UpdateTransferV2To(tx types.ITransaction, blockHeight uint64) error

	UpdateTransferTo(tx types.ITransaction, blockHeight uint64) error

	UpdateContractTo(tx types.ITransaction, blockHeight uint64) error

	UpdateFees(fees, blockHeight uint64) error

	UpdateConsumption(consumption, blockHeight uint64) error

	UpdateConfirmedHeight(height uint64)

	VerifyState(tx types.ITransaction) error

	StateTrieCommit() (hasharry.Hash, error)

	RootHash() hasharry.Hash

	Print()

	Close() error
}
