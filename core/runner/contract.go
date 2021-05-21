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

func (c *ContractRunner) Verify(tx types.ITransaction, lastHeight uint64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if tx.GetTxType() != types.ContractV2_ {
		return nil
	}
	body, _ := tx.GetTxBody().(*types.TxContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library, tx)
		switch body.FunctionType {
		case contractv2.Exchange_Init:
			return ex.PreInitVerify()
		case contractv2.Exchange_SetAdmin:
			return ex.PreSetVerify()
		case contractv2.Exchange_SetFeeTo:
			return ex.PreSetVerify()
		case contractv2.Exchange_ExactIn:
			return ex.PreExactInVerify(lastHeight)
		case contractv2.Exchange_ExactOut:
			return ex.PreExactOutVerify(lastHeight)
		}

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

	body, _ := tx.GetTxBody().(*types.TxContractV2Body)
	switch body.Type {
	case contractv2.Exchange_:
		ex := exchange_runner.NewExchangeRunner(c.library, tx)
		switch body.FunctionType {
		case contractv2.Exchange_Init:
			ex.Init()
		case contractv2.Exchange_SetAdmin:
			ex.SetAdmin()
		case contractv2.Exchange_SetFeeTo:
			ex.SetFeeTo()
		case contractv2.Exchange_ExactIn:
			ex.SwapExactIn(blockHeight, blockTime)
		case contractv2.Exchange_ExactOut:
			ex.SwapExactOut(blockHeight, blockTime)
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
