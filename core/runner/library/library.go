package library

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

type RunnerLibrary struct {
	aState _interface.IAccountState
	cState _interface.IContractState
}

func NewRunnerLibrary(aState _interface.IAccountState, cState _interface.IContractState) *RunnerLibrary {
	return &RunnerLibrary{aState: aState, cState: cState}
}

func (r *RunnerLibrary) GetContract(contractAddr string) *types.Contract {
	return r.cState.GetContract(contractAddr)
}

func (r *RunnerLibrary) SetContract(contract *types.Contract) {
	r.cState.SetContract(contract)
}

func (r *RunnerLibrary) GetContractV2(contractAddr string) *contractv2.ContractV2 {
	return r.cState.GetContractV2(contractAddr)
}

func (r *RunnerLibrary) SetContractV2(contract *contractv2.ContractV2) {
	r.cState.SetContractV2(contract)
}

func (r *RunnerLibrary) Transfer(sender, to, token hasharry.Address, amount, height uint64) error {
	return r.aState.Transfer(sender, to, token, amount, height)
}
