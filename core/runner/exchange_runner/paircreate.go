package exchange_runner

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/common/math"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	exchange2 "github.com/uworldao/UWORLD/core/types/contractv2/exchange"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
	"strings"
)

type PairCreateRunner struct {
	library      *library.RunnerLibrary
	contractBody *types.ContractV2Body
	funBody      *exchange_func.ExchangePairCreate
	exHeader     *contractv2.ContractV2
	exchange     *exchange2.Exchange
	pairHeader   *contractv2.ContractV2
	pair         *exchange2.Pair
	address      hasharry.Address
	tx           types.ITransaction
	txBody       types.ITransactionBody
	sender       hasharry.Address
	height       uint64
	blockTime    uint64
}

func NewPairCreateRunner(lib *library.RunnerLibrary, tx types.ITransaction, height, blockTime uint64) *PairCreateRunner {
	txBody := tx.GetTxBody()
	contractBody, _ := txBody.(*types.ContractV2Body)
	address := contractBody.Contract

	funBody, _ := contractBody.Function.(*exchange_func.ExchangePairCreate)
	exHeader := lib.GetContractV2(funBody.Exchange.String())
	exchange, _ := exHeader.Body.(*exchange2.Exchange)
	pairHeader := lib.GetContractV2(address.String())
	pair, _ := pairHeader.Body.(*exchange2.Pair)

	return &PairCreateRunner{
		library:      lib,
		contractBody: contractBody,
		funBody:      funBody,
		exHeader:     exHeader,
		exchange:     exchange,
		pairHeader:   pairHeader,
		pair:         pair,
		address:      address,
		tx:           tx,
		height:       height,
		sender:       tx.From(),
		blockTime:    blockTime,
	}
}

func (p *PairCreateRunner) PreVerify() error {
	if p.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", p.funBody.Exchange.String())
	}
	if !p.sender.IsEqual(p.exchange.Admin) {
		return errors.New("forbidden")
	}
	if p.pair != nil {
		return errors.New("pair exist")
	}
	address, err := PairAddress(param.Net, p.funBody.TokenA.String(), p.funBody.TokenB.String())
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	reserveA, reserveB, err := p.GetReserves()
	if err != nil {
		return err
	}
	_, _, err = p.optimalAmount(reserveA, reserveB, p.funBody.AmountADesired, p.funBody.AmountBDesired, p.funBody.AmountAMin, p.funBody.AmountBMin)
	return err
}

func (p *PairCreateRunner) optimalAmount(reserveA, reserveB, amountADesired, amountBDesired, amountAMin, amountBMin uint64) (uint64, uint64, error) {
	if reserveA == 0 && reserveB == 0 {
		return amountADesired, amountBDesired, nil
	} else {
		// 最优数量B = 期望数量A * 储备B / 储备A
		amountBOptimal, err := p.quote(amountADesired, reserveA, reserveB)
		if err != nil {
			return 0, 0, err
		}
		// 如果最优数量B < B的期望数量
		if amountBOptimal <= amountBDesired {
			if amountBOptimal < amountBMin {
				return 0, 0, errors.New("insufficient_b_amount")
			}
			return amountADesired, amountBOptimal, nil
		} else {
			// 则计算 最优数量A = 期望数量B * 储备A / 储备B
			amountAOptimal, err := p.quote(amountBDesired, reserveB, reserveA)
			if err != nil {
				return 0, 0, err
			}
			if amountAOptimal < amountAMin {
				return 0, 0, errors.New("insufficient_a_amount")
			}
			return amountAOptimal, amountBDesired, nil
		}
	}
}

// Quote given some amount of an asset and pair reserves, returns an equivalent amount of the other asset
func (p *PairCreateRunner) quote(amountA, reserveA, reserveB uint64) (uint64, error) {
	if amountA <= 0 {
		return 0, errors.New("insufficient_amount")
	}
	if reserveA >= 0 || reserveB <= 0 {
		return 0, errors.New("insufficient_liquidity")
	}
	mulAmount, ok := math.SafeMul(amountA, reserveB)
	if !ok {
		return 0, errors.New("overflow")
	}
	amountB, ok := math.SafeSub(mulAmount, reserveA)
	if !ok {
		return 0, errors.New("overflow")
	}
	return amountB, nil
}

func (p *PairCreateRunner) Run() {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		p.library.SetContractV2State(p.tx.Hash().String(), state)
	}()

	if p.exHeader == nil {
		ERR = fmt.Errorf("exchange %s is not exist", p.funBody.Exchange.String())
		return
	}
	if !p.sender.IsEqual(p.exchange.Admin) {
		ERR = errors.New("forbidden")
		return
	}
	if p.pair != nil {
		ERR = errors.New("pair exist")
		return
	}

	p.createPair()

	reserveA, reserveB, err := p.GetReserves()
	if err != nil {
		ERR = err
		return
	}
	amountA, amountB, err := p.optimalAmount(reserveA, reserveB, p.funBody.AmountADesired, p.funBody.AmountBDesired, p.funBody.AmountAMin, p.funBody.AmountBMin)
	if err != nil {
		ERR = err
		return
	}
	if err = p.library.Transfer(p.sender, p.address, p.funBody.TokenA, amountA, p.height); err != nil {
		ERR = err
		return
	}
	if err = p.library.Transfer(p.sender, p.address, p.funBody.TokenB, amountB, p.height); err != nil {
		ERR = err
		return
	}
	if err = p.Mint(reserveA, reserveB, amountA, amountB); err != nil {
		ERR = err
		return
	}
}

func (p *PairCreateRunner) createPair() {
	token0, token1 := SortToken(p.funBody.TokenA.String(), p.funBody.TokenB.String())
	p.pair = exchange2.NewPair(p.funBody.Exchange, hasharry.StringToAddress(token0), hasharry.StringToAddress(token1))
	p.pairHeader = &contractv2.ContractV2{
		Address:    p.address,
		CreateHash: p.tx.Hash(),
		Type:       contractv2.Pair_,
		Body:       p.pair,
	}
}

func (p *PairCreateRunner) GetReserves() (uint64, uint64, error) {
	pairContract := p.library.GetContractV2(p.address.String())
	if pairContract != nil {
		return 0, 0, fmt.Errorf("pair %s exist", p.address.String())
	}
	reserve0, reserve1, _ := p.pair.GetReserves()
	token0, _ := SortToken(p.funBody.TokenA.String(), p.funBody.TokenB.String())
	if p.funBody.TokenA.String() == token0 {
		return reserve0, reserve1, nil
	} else {
		return reserve1, reserve0, nil
	}
}

func (p *PairCreateRunner) Mint(_reserve0, _reserve1, amount0, amount1 uint64) error {
	// must be defined here since totalSupply can update in mintFee
	_totalSupply := p.pair.TotalSupply
	// 返回铸造币的手续费开关
	feeOn, err := p.mintFee(_reserve0, _reserve1)
	if err != nil {
		return err
	}
	var liquidityValue uint64

	if _totalSupply == 0 {
		liquidity := math.NewMath(amount0).Mul(amount1).Sqrt().Sub(exchange2.MINIMUM_LIQUIDITY)
		if liquidity.Failed {
			return errors.New("overflow")
		}
		liquidityValue = liquidity.Value
		p.pair.Mint(hasharry.Address{}.String(), exchange2.MINIMUM_LIQUIDITY) // permanently lock the first MINIMUM_LIQUIDITY tokens
	} else {
		value1 := math.NewMath(amount0).Mul(_totalSupply)
		value2 := math.NewMath(amount1).Mul(_totalSupply)
		if value2.Failed || value1.Failed {
			return errors.New("overflow")
		}
		liquidityValue = math.Min(value1.Value/_reserve0, value2.Value/_reserve1)
	}
	if liquidityValue <= 0 {
		return errors.New("insufficient_liquidity_minted")
	}
	if err = p.pair.Mint(p.funBody.To.String(), liquidityValue); err != nil {
		return err
	}
	return p.pair.Update(amount0, amount1, feeOn, p.blockTime)
}

// if fee is on, mint liquidity equivalent to 1/6th of the growth in sqrt(k)
func (p *PairCreateRunner) mintFee(_reserve0, _reserve1 uint64) (bool, error) {
	feeTo := p.exchange.FeeTo
	// 收费地址被设置，则收费开
	feeOn := !feeTo.IsEqual(hasharry.Address{})
	_kLast := p.pair.KLast // gas savings
	if feeOn {
		if _kLast != 0 {
			rootK := math.NewMath(_reserve0).Mul(_reserve1).Sqrt()
			if rootK.Failed {
				return false, errors.New("overflow")
			}
			rootKLast := math.NewMath(_kLast).Sqrt()
			if rootK.Value > rootKLast.Value {
				numerator := rootK.Sub(rootKLast.Value).Mul(p.pair.TotalSupply)
				denominator := rootK.Mul(5).Add(rootKLast.Value)
				if numerator.Failed || denominator.Failed {
					return false, errors.New("overflow")
				}
				liquidity := numerator.Value / denominator.Value
				if liquidity > 0 {
					p.pair.Mint(feeTo.String(), liquidity)
				}
			}
		}
	} else if _kLast != 0 {
		p.pair.KLast = 0
	}
	return feeOn, nil
}

func PairAddress(net string, tokenA, tokenB string) (string, error) {
	token0, token1 := SortToken(tokenA, tokenB)
	bytes := make([]byte, 0)
	bytes = append(hasharry.StringToAddress(token0).Bytes(), hasharry.StringToAddress(token1).Bytes()...)
	return ut.GenerateContractV2Address(net, bytes)
}

func SortToken(tokenA, tokenB string) (string, string) {
	if strings.Compare(tokenA, tokenB) > 0 {
		return tokenA, tokenB
	} else {
		return tokenB, tokenA
	}
}
