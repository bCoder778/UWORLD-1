package runner

import (
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/runner/exchange_runner"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
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
		ex := exchange_runner.NewExchangeRunner(c.library, tx, lastHeight)
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
		case contractv2.Pair_AddLiquidity:
			pair := exchange_runner.NewPairRunner(c.library, tx, 0, 0)
			return pair.PreAddLiquidityVerify()
		case contractv2.Pair_RemoveLiquidity:
			pair := exchange_runner.NewPairRunner(c.library, tx, 0, 0)
			return pair.PreRemoveLiquidityVerify(lastHeight)
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
		ex := exchange_runner.NewExchangeRunner(c.library, tx, blockHeight)
		switch body.FunctionType {
		case contractv2.Exchange_Init:
			ex.Init()
		case contractv2.Exchange_SetAdmin:
			ex.SetAdmin()
		case contractv2.Exchange_SetFeeTo:
			ex.SetFeeTo()
		case contractv2.Exchange_ExactIn:
			ex.SwapExactIn(blockTime)
		case contractv2.Exchange_ExactOut:
			ex.SwapExactOut(blockTime)
		}
	case contractv2.Pair_:
		switch body.FunctionType {
		case contractv2.Pair_AddLiquidity:
			pairRunner := exchange_runner.NewPairRunner(c.library, tx, blockHeight, blockTime)
			pairRunner.AddLiquidity()
		case contractv2.Pair_RemoveLiquidity:
			pairRunner := exchange_runner.NewPairRunner(c.library, tx, blockHeight, blockTime)
			pairRunner.RemoveLiquidity()
		}
	}
	return nil
}

func (c *ContractRunner) ExchangePair(address hasharry.Address) ([]*types.RpcPair, error) {
	exHeader := c.library.GetContractV2(address.String())
	if exHeader == nil {
		return nil, fmt.Errorf("exchange %s is not exist", address.String())
	}
	rpcPairList := make([]*types.RpcPair, 0)
	ex := exHeader.Body.(*exchange.Exchange)
	for _, pair := range ex.AllPairs {
		token0, token1 := exchange.ParseKey(pair.Key)
		rpcPairList = append(rpcPairList, &types.RpcPair{
			Address:  pair.Address.String(),
			Token0:   token0.String(),
			Token1:   token1.String(),
			Reserve0: c.library.GetBalance(pair.Address, token0),
			Reserve1: c.library.GetBalance(pair.Address, token1),
		})
	}
	return rpcPairList, nil
}
