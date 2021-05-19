package _interface

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

type IContractState interface {
	GetContract(contractAddr string) *types.Contract

	SetContract(contract *types.Contract)

	GetContractV2(contractAddr string) *contractv2.ContractV2

	SetContractV2(contract *contractv2.ContractV2)

	VerifyState(tx types.ITransaction) error

	UpdateContract(tx types.ITransaction, blockHeight uint64)

	UpdateConfirmedHeight(height uint64)

	InitTrie(hash hasharry.Hash) error

	RootHash() hasharry.Hash

	ContractTrieCommit() (hasharry.Hash, error)

	Close() error
}
