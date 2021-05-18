package contractstate

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

// Implement storage as contract state
type IContractStorage interface {
	GetContractState(contractAddr string) *types.Contract
	SetContractState(contract *types.Contract)
	GetContractV2State(contractAddr string) *contractv2.ContractV2
	SetContractV2State(contract *contractv2.ContractV2)
	InitTrie(contractRoot hasharry.Hash) error
	RootHash() hasharry.Hash
	Commit() (hasharry.Hash, error)
	Close() error
}
