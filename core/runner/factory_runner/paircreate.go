package factory_runner

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/common/math"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	factory2 "github.com/uworldao/UWORLD/core/types/contractv2/factory"
	"github.com/uworldao/UWORLD/core/types/functionbody/factory_func"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
	"math/big"
	"strings"
)

type PairRunner struct {
	library      *library.RunnerLibrary
	contractBody *types.ContractV2Body
	funBody      *factory_func.FactoryPairCreate
	exHeader     *contractv2.ContractV2
	factory      *factory2.Factory
	pairHeader   *contractv2.ContractV2
	pair         *factory2.Pair
	address      hasharry.Address
	tx           types.ITransaction
	txBody       types.ITransactionBody
	sender       hasharry.Address
	state        *types.ContractV2State
	height       uint64
	blockTime    uint64
}

func NewPairRunner(lib *library.RunnerLibrary, tx types.ITransaction, height, blockTime uint64) *PairRunner {
	var factory *factory2.Factory
	var pair *factory2.Pair
	txBody := tx.GetTxBody()
	contractBody, _ := txBody.(*types.ContractV2Body)
	address := contractBody.Contract

	funBody, _ := contractBody.Function.(*factory_func.FactoryPairCreate)
	exHeader := lib.GetContractV2(funBody.Factory.String())
	if exHeader != nil {
		factory, _ = exHeader.Body.(*factory2.Factory)
	}

	pairHeader := lib.GetContractV2(address.String())
	if pairHeader != nil {
		pair, _ = pairHeader.Body.(*factory2.Pair)
	}
	state := &types.ContractV2State{State: types.Contract_Success}
	return &PairRunner{
		library:      lib,
		contractBody: contractBody,
		funBody:      funBody,
		exHeader:     exHeader,
		factory:      factory,
		pairHeader:   pairHeader,
		pair:         pair,
		address:      address,
		state:        state,
		tx:           tx,
		height:       height,
		sender:       tx.From(),
		blockTime:    blockTime,
	}
}

func (p *PairRunner) PreCreateVerify() error {
	if p.exHeader == nil {
		return fmt.Errorf("factory %s is not exist", p.funBody.Factory.String())
	}
	if !p.sender.IsEqual(p.factory.Admin) {
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
	return nil
}

func (p *PairRunner) PreAddVerify() error {
	if p.exHeader == nil {
		return fmt.Errorf("factory %s is not exist", p.funBody.Factory.String())
	}
	if !p.sender.IsEqual(p.factory.Admin) {
		return errors.New("forbidden")
	}
	if p.pair == nil {
		return errors.New("pair not exist")
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

func (p *PairRunner) optimalAmount(reserveA, reserveB, amountADesired, amountBDesired, amountAMin, amountBMin uint64) (uint64, uint64, error) {
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
func (p *PairRunner) quote(amountA, reserveA, reserveB uint64) (uint64, error) {
	if amountA <= 0 {
		return 0, errors.New("insufficient_amount")
	}
	if reserveA >= 0 || reserveB <= 0 {
		return 0, errors.New("insufficient_liquidity")
	}
	amountB := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(int64(amountA)), big.NewInt(int64(reserveB))), big.NewInt(int64(reserveA)))
	return amountB.Uint64(), nil
}

func (p *PairRunner) Create() {
	var ERR error
	defer func() {
		if ERR != nil {
			p.state.State = types.Contract_Failed
			p.state.Message = ERR.Error()
		}
		p.update()
	}()

	if p.exHeader == nil {
		ERR = fmt.Errorf("factory %s is not exist", p.funBody.Factory.String())
		return
	}
	if !p.sender.IsEqual(p.factory.Admin) {
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
	if err = p.mint(reserveA, reserveB, amountA, amountB); err != nil {
		ERR = err
		return
	}
}

func (p *PairRunner) createPair() {
	token0, token1 := SortToken(p.funBody.TokenA.String(), p.funBody.TokenB.String())
	p.pair = factory2.NewPair(p.funBody.Factory, hasharry.StringToAddress(token0), hasharry.StringToAddress(token1))
	p.pairHeader = &contractv2.ContractV2{
		Address:    p.address,
		CreateHash: p.tx.Hash(),
		Type:       contractv2.Pair_,
		Body:       p.pair,
	}
	p.factory.AddPair(hasharry.StringToAddress(token0), hasharry.StringToAddress(token1), p.address)
}

func (p *PairRunner) GetReserves() (uint64, uint64, error) {
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

func (p *PairRunner) mint(_reserve0, _reserve1, amount0, amount1 uint64) error {
	// must be defined here since totalSupply can update in mintFee
	_totalSupply := p.pair.TotalSupply
	// 返回铸造币的手续费开关
	feeOn, err := p.mintFee(_reserve0, _reserve1)
	if err != nil {
		return err
	}
	var liquidityValue uint64

	if _totalSupply == 0 {
		liquidityBig := big.NewInt(0).Sub(big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(amount1)))), big.NewInt(int64(factory2.MINIMUM_LIQUIDITY)))
		liquidityValue = liquidityBig.Uint64()
		p.pair.Mint(hasharry.Address{}.String(), factory2.MINIMUM_LIQUIDITY) // permanently lock the first MINIMUM_LIQUIDITY tokens
	} else {
		value1 := big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(_totalSupply)))
		value2 := big.NewInt(0).Mul(big.NewInt(int64(amount1)), big.NewInt(int64(_totalSupply)))
		liquidityValue = math.Min(value1.Uint64()/_reserve0, value2.Uint64()/_reserve1)
	}
	if liquidityValue <= 0 {
		return errors.New("insufficient_liquidity_minted")
	}
	p.pair.Mint(p.funBody.To.String(), liquidityValue)
	p.pair.Update(amount0, amount1, feeOn, p.blockTime)
	return nil
}

// if fee is on, mint liquidity equivalent to 1/6th of the growth in sqrt(k)
func (p *PairRunner) mintFee(_reserve0, _reserve1 uint64) (bool, error) {
	feeTo := p.factory.FeeTo
	// 收费地址被设置，则收费开
	feeOn := !feeTo.IsEqual(hasharry.Address{})
	_kLast := p.pair.KLast // gas savings
	if feeOn {
		if _kLast.Cmp(big.NewInt(0)) != 0 {
			// roottK = Sqrt(_reserve0 * _reserve1)
			rootK := big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(_reserve0)), big.NewInt(int64(_reserve1))))
			// rootKLast = Sqrt(_kLast)
			rootKLast := big.NewInt(0).Sqrt(_kLast)
			if rootK.Cmp(rootKLast) > 0 {
				// numerator = (rootK-rootKLast)*TotalSupply
				numerator := big.NewInt(0).Mul(big.NewInt(0).Sub(rootK, rootKLast), big.NewInt(int64(p.pair.TotalSupply)))
				// denominator = rootK * 5 + rootKLast
				denominator := big.NewInt(0).Add(big.NewInt(0).Mul(rootK, big.NewInt(5)), rootKLast)
				// liquidity = numerator / denominator
				liquidityBig := big.NewInt(0).Div(numerator, denominator)
				if liquidityBig.Cmp(big.NewInt(0)) > 0 {
					p.pair.Mint(feeTo.String(), liquidityBig.Uint64())
				}
			}
		}
	} else if _kLast.Cmp(big.NewInt(0)) != 0 {
		p.pair.KLast = big.NewInt(0)
	}
	return feeOn, nil
}

func (p *PairRunner) update() {
	p.exHeader.Body = p.factory
	p.pairHeader.Body = p.pair
	p.library.SetContractV2(p.exHeader)
	p.library.SetContractV2(p.pairHeader)
	p.library.SetContractV2State(p.tx.Hash().String(), p.state)
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
