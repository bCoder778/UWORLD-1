package exchange_func

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type ExchangeAddLiquidity struct {
	Exchange       hasharry.Address
	TokenA         hasharry.Address
	TokenB         hasharry.Address
	To             hasharry.Address
	AmountADesired uint64
	AmountBDesired uint64
	AmountAMin     uint64
	AmountBMin     uint64
}

func (e *ExchangeAddLiquidity) Verify() error {
	if e.TokenA.IsEqual(e.TokenB) {
		return errors.New("invalid pair")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.TokenA.String()); !ok {
		return errors.New("wrong tokenA address")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.TokenB.String()); !ok {
		return errors.New("wrong tokenB address")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.Exchange.String()); !ok {
		return errors.New("wrong exchange address")
	}
	if ok := ut.CheckUWDAddress(param.Net, e.To.String()); !ok {
		return errors.New("wrong to address")
	}
	if e.AmountADesired == 0 {
		return errors.New("wrong amountADesired")
	}
	if e.AmountBDesired == 0 {
		return errors.New("wrong amountBDesired")
	}
	if e.AmountAMin > e.AmountADesired {
		return errors.New("wrong amountAMin")
	}
	if e.AmountBMin > e.AmountBDesired {
		return errors.New("wrong amountBMin")
	}
	return nil
}

type ExchangeRemoveLiquidity struct {
	Exchange   hasharry.Address
	TokenA     hasharry.Address
	TokenB     hasharry.Address
	To         hasharry.Address
	Liquidity  uint64
	AmountAMin uint64
	AmountBMin uint64
	Deadline   uint64
}

func (e *ExchangeRemoveLiquidity) Verify() error {
	if e.TokenA.IsEqual(e.TokenB) {
		return errors.New("invalid pair")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.TokenA.String()); !ok {
		return errors.New("wrong tokenA address")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.TokenB.String()); !ok {
		return errors.New("wrong tokenB address")
	}
	if ok := ut.IsValidContractAddress(param.Net, e.Exchange.String()); !ok {
		return errors.New("wrong exchange address")
	}
	if ok := ut.CheckUWDAddress(param.Net, e.To.String()); !ok {
		return errors.New("wrong to address")
	}
	return nil
}
