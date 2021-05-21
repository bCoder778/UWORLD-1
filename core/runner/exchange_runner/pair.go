package exchange_runner

import (
	bytes2 "bytes"
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
	"math/big"
)

type PairRunner struct {
	library      *library.RunnerLibrary
	contractBody *types.TxContractV2Body
	funBody      *exchange_func.ExchangePairCreate
	exHeader     *contractv2.ContractV2
	exchange     *exchange2.Exchange
	pairHeader   *contractv2.ContractV2
	pair         *exchange2.Pair
	address      hasharry.Address
	tx           types.ITransaction
	txBody       types.ITransactionBody
	sender       hasharry.Address
	state        *types.ContractV2State
	height       uint64
	blockTime    uint64
}

func NewPairRunner(lib *library.RunnerLibrary, tx types.ITransaction, height, blockTime uint64) *PairRunner {
	var exchange *exchange2.Exchange
	var pair *exchange2.Pair
	txBody := tx.GetTxBody()
	contractBody, _ := txBody.(*types.TxContractV2Body)
	address := contractBody.Contract

	funBody, _ := contractBody.Function.(*exchange_func.ExchangePairCreate)
	exHeader := lib.GetContractV2(funBody.Exchange.String())
	if exHeader != nil {
		exchange, _ = exHeader.Body.(*exchange2.Exchange)
	}

	pairHeader := lib.GetContractV2(address.String())
	if pairHeader != nil {
		pair, _ = pairHeader.Body.(*exchange2.Pair)
	}
	state := &types.ContractV2State{State: types.Contract_Success}
	return &PairRunner{
		library:      lib,
		contractBody: contractBody,
		funBody:      funBody,
		exHeader:     exHeader,
		exchange:     exchange,
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
		return fmt.Errorf("exchange %s is not exist", p.funBody.Exchange.String())
	}
	if !p.sender.IsEqual(p.exchange.Admin) {
		return errors.New("forbidden")
	}
	if p.pair != nil {
		return errors.New("pair exist")
	}
	address, err := PairAddress(param.Net, p.funBody.TokenA, p.funBody.TokenB, p.exHeader.Address)
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
		return fmt.Errorf("exchange %s is not exist", p.funBody.Exchange.String())
	}
	if !p.sender.IsEqual(p.exchange.Admin) {
		return errors.New("forbidden")
	}
	if p.pair == nil {
		return errors.New("pair not exist")
	}
	address, err := PairAddress(param.Net, p.funBody.TokenA, p.funBody.TokenB, p.exHeader.Address)
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	pairContract := p.library.GetContractV2(address)
	if pairContract == nil {
		return fmt.Errorf("the pair %s is not exist", address)
	}
	pair := pairContract.Body.(*exchange2.Pair)
	reserveA, reserveB := p.library.GetReservesByPair(pair, p.funBody.TokenA, p.funBody.TokenB)
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

	reserveA, reserveB := p.library.GetReservesByPair(p.pair, p.funBody.TokenA, p.funBody.TokenB)

	amountA, amountB, err := p.optimalAmount(reserveA, reserveB, p.funBody.AmountADesired, p.funBody.AmountBDesired, p.funBody.AmountAMin, p.funBody.AmountBMin)
	if err != nil {
		ERR = err
		return
	}
	transInfo1 := &library.TransferInfo{
		From:   p.sender,
		To:     p.address,
		Token:  p.funBody.TokenA,
		Amount: amountA,
		Height: p.height,
	}
	if err = p.library.PreTransfer(transInfo1); err != nil {
		ERR = err
		return
	}
	transInfo2 := &library.TransferInfo{
		From:   p.sender,
		To:     p.address,
		Token:  p.funBody.TokenB,
		Amount: amountB,
		Height: p.height,
	}
	if err = p.library.PreTransfer(transInfo2); err != nil {
		ERR = err
		return
	}
	if err = p.mint(reserveA, reserveB, amountA, amountB); err != nil {
		ERR = err
		return
	}
	p.library.Transfer(transInfo1)
	p.library.Transfer(transInfo2)
	p.update()
}

func (p *PairRunner) createPair() {
	token0, token1 := library.SortToken(p.funBody.TokenA, p.funBody.TokenB)
	p.pair = exchange2.NewPair(p.funBody.Exchange, token0, token1)
	p.pairHeader = &contractv2.ContractV2{
		Address:    p.address,
		CreateHash: p.tx.Hash(),
		Type:       contractv2.Pair_,
		Body:       p.pair,
	}
	p.exchange.AddPair(token0, token1, p.address)
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
		liquidityBig := big.NewInt(0).Sub(big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(amount1)))), big.NewInt(int64(exchange2.MINIMUM_LIQUIDITY)))
		liquidityValue = liquidityBig.Uint64()
		p.pair.Mint(hasharry.Address{}.String(), exchange2.MINIMUM_LIQUIDITY) // permanently lock the first MINIMUM_LIQUIDITY tokens
	} else {
		value1 := big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(_totalSupply)))
		value2 := big.NewInt(0).Mul(big.NewInt(int64(amount1)), big.NewInt(int64(_totalSupply)))
		liquidityValue = math.Min(value1.Uint64()/_reserve0, value2.Uint64()/_reserve1)
	}
	if liquidityValue <= 0 {
		return errors.New("insufficient_liquidity_minted")
	}
	p.pair.Mint(p.funBody.To.String(), liquidityValue)
	p.pair.Update(_reserve0+amount0, _reserve1+amount1, _reserve0, _reserve1, p.blockTime)
	if feeOn {
		p.pair.UpdateKLast()
	}
	return nil
}

// if fee is on, mint liquidity equivalent to 1/6th of the growth in sqrt(k)
func (p *PairRunner) mintFee(_reserve0, _reserve1 uint64) (bool, error) {
	feeTo := p.exchange.FeeTo
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
	p.exHeader.Body = p.exchange
	p.pairHeader.Body = p.pair
	p.library.SetContractV2(p.exHeader)
	p.library.SetContractV2(p.pairHeader)
	p.library.SetContractV2State(p.tx.Hash().String(), p.state)
}

func PairAddress(net string, tokenA, tokenB hasharry.Address, exchange hasharry.Address) (string, error) {
	token0, token1 := library.SortToken(tokenA, tokenB)
	bytes := bytes2.Join([][]byte{[]byte(token0.String()), []byte(token1.String()), []byte(exchange.String())}, []byte{})
	return ut.GenerateContractV2Address(net, bytes)
}
