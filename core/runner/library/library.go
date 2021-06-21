package library

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
	"strings"
)

type RunnerLibrary struct {
	aState _interface.IAccountState
	cState _interface.IContractState
}

func NewRunnerLibrary(aState _interface.IAccountState, cState _interface.IContractState) *RunnerLibrary {
	return &RunnerLibrary{aState: aState, cState: cState}
}

func (r *RunnerLibrary) CreateToken(height uint64, hash hasharry.Hash, time uint64, from hasharry.Address, token0, token1 hasharry.Address) (hasharry.Address, error) {
	token0Record := r.cState.GetContract(token0.String())
	token1Record := r.cState.GetContract(token1.String())
	abbr := fmt.Sprintf("LP-%s-%s", token0Record.CoinAbbr, token1Record.CoinAbbr)
	lp, err := ut.GenerateContractAddress(param.Net, from.String(), abbr)
	if err != nil {
		return hasharry.Address{}, err
	}
	lpAddress := hasharry.StringToAddress(lp)
	r.cState.CreateTokenContract(lpAddress, abbr, "", abbr, hash, height, time, 0, from, true)
	return lpAddress, nil
}

func (r *RunnerLibrary) Mint(contract hasharry.Address, hash hasharry.Hash, receiver hasharry.Address, amount, height, time uint64) {
	_ = r.cState.IncreaseTokenContract(contract, hash, height, time, amount, receiver)
	_ = r.aState.Mint(receiver, contract, amount, height)
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

func (r *RunnerLibrary) PreTransfer(info *TransferInfo) error {
	return r.aState.PreTransfer(info.From, info.To, info.Token, info.Amount, info.Height)
}

func (r *RunnerLibrary) Transfer(info *TransferInfo) error {
	return r.aState.Transfer(info.From, info.To, info.Token, info.Amount, info.Height)
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

type TransferInfo struct {
	From   hasharry.Address
	To     hasharry.Address
	Token  hasharry.Address
	Amount uint64
	Height uint64
}
