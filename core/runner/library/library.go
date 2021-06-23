package library

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
	"strings"
)

type RunnerLibrary struct {
	aState _interface.IAccountState
	cState _interface.IContractState
}

func NewRunnerLibrary(aState _interface.IAccountState, cState _interface.IContractState) *RunnerLibrary {
	return &RunnerLibrary{aState: aState, cState: cState}
}

func (r *RunnerLibrary) ContractSymbol(token hasharry.Address) (string, error) {
	token0Record := r.cState.GetContract(token.String())
	if token0Record == nil {
		return "", fmt.Errorf("%s is not exist", token.String())
	}
	return token0Record.CoinAbbr, nil
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

func (r RunnerLibrary) SetContractV2State(txHash string, state *types.ContractV2State) {
	r.cState.SetContractV2State(txHash, state)
}

func (r RunnerLibrary) GetBalance(address hasharry.Address, token hasharry.Address) uint64 {
	account := r.aState.GetAccountState(address)
	return account.GetBalance(token.String())
}

func (r *RunnerLibrary) PreRunEvent(event *types.Event) error {
	switch event.EventType {
	case types.Event_Transfer:
		return r.aState.PreTransfer(event.From, event.To, event.Token, event.Amount, event.Height)
	case types.Event_Mint:
		return nil
	case types.Event_Burn:
		return r.aState.PreBurn(event.From, event.Token, event.Amount, event.Height)
	}
	return fmt.Errorf("invalid event type")
}

func (r *RunnerLibrary) RunEvent(event *types.Event) {
	switch event.EventType {
	case types.Event_Transfer:
		r.aState.Transfer(event.From, event.To, event.Token, event.Amount, event.Height)
	case types.Event_Mint:
		r.aState.Mint(event.To, event.Token, event.Amount, event.Height)
	case types.Event_Burn:
		r.aState.Burn(event.From, event.Token, event.Amount, event.Height)
	}
}

func (r *RunnerLibrary) GetPair(pairAddress hasharry.Address) (*exchange.Pair, error) {
	pairContract := r.GetContractV2(pairAddress.String())
	if pairContract != nil {
		return nil, errors.New("%s pair does not exist")
	}
	return pairContract.Body.(*exchange.Pair), nil
}

func (r *RunnerLibrary) GetExchange(exchangeAddress hasharry.Address) (*exchange.Exchange, error) {
	exContract := r.GetContractV2(exchangeAddress.String())
	if exContract != nil {
		return nil, errors.New("%s exchange does not exist")
	}
	return exContract.Body.(*exchange.Exchange), nil
}

func (r *RunnerLibrary) GetReservesByPairAddress(pairAddress, tokenA, tokenB hasharry.Address) (uint64, uint64) {
	pairContract := r.GetContractV2(pairAddress.String())
	pair := pairContract.Body.(*exchange.Pair)
	return r.GetReservesByPair(pair, tokenA, tokenB)
}

func (r *RunnerLibrary) GetReservesByPair(pair *exchange.Pair, tokenA, tokenB hasharry.Address) (uint64, uint64) {
	reserve0, reserve1, _ := pair.GetReserves()
	token0, _ := SortToken(tokenA, tokenB)
	if tokenA.IsEqual(token0) {
		return reserve0, reserve1
	} else {
		return reserve1, reserve0
	}
}

func SortToken(tokenA, tokenB hasharry.Address) (hasharry.Address, hasharry.Address) {
	if strings.Compare(tokenA.String(), tokenB.String()) > 0 {
		return tokenA, tokenB
	} else {
		return tokenB, tokenA
	}
}
