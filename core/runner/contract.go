package runner

import (
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/runner/exchange_runner"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"sync"
)

type ContractRunner struct {
	mutex   sync.RWMutex
	library *library.RunnerLibrary
}

func NewContractRunner(accountState _interface.IAccountState, contractState _interface.IContractState) *ContractRunner {
	library := library.NewRunnerLibrary(accountState, contractState)
	return &ContractRunner{
		library: library,
	}
}

func (c *ContractRunner) Verify(tx types.ITransaction) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if tx.GetTxType() != types.ContractV2_ {
		return nil
	}
	body, _ := tx.GetTxBody().(*types.ContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library)
		return ex.PreVerify(tx.From(), body.Contract, body.FunctionType)
	case contractv2.Pair_:
		switch body.FunctionType {
		case contractv2.Pair_Create:
			pair := exchange_runner.NewPairRunner(c.library, tx, 0, 0)
			return pair.PreCreateVerify()
		}
	}
	return nil
}

func (c *ContractRunner) RunContract(tx types.ITransaction, blockHeight uint64, blockTime uint64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	body, _ := tx.GetTxBody().(*types.ContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library)
		switch body.FunctionType {
		case contractv2.Exchange_Init_:
			ex.Init(tx.GetTxHead(), body, blockHeight)
		case contractv2.Exchange_SetAdmin_:
			ex.SetAdmin(tx.GetTxHead(), body, blockHeight)
		case contractv2.Exchange_SetFeeTo_:
			ex.SetFeeTo(tx.GetTxHead(), body, blockHeight)
		}
	case contractv2.Pair_:
		switch body.FunctionType {
		case contractv2.Pair_Create:
			pairRunner := exchange_runner.NewPairRunner(c.library, tx, blockHeight, blockTime)
			pairRunner.Create()
		}
	}
	return nil
}
