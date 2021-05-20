package contractstate

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

// Implement storage as contract state
type IContractStorage interface {
	GetContract(contractAddr string) *types.Contract
	SetContract(contract *types.Contract)
	GetContractV2(contractAddr string) *contractv2.ContractV2
	SetContractV2(contract *contractv2.ContractV2)
	SetContractV2State(txHash string, state *types.ContractV2State)
	GetContractV2State(txHash string) *types.ContractV2State
	InitTrie(contractRoot hasharry.Hash) error
	RootHash() hasharry.Hash
	Commit() (hasharry.Hash, error)
	Close() error
}
