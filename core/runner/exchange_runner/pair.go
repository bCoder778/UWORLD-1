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
	addBody      *exchange_func.ExchangeAddLiquidity
	removeBody   *exchange_func.ExchangeRemoveLiquidity
	exHeader     *contractv2.ContractV2
	exchange     *exchange2.Exchange
	pairHeader   *contractv2.ContractV2
	pair         *exchange2.Pair
	address      hasharry.Address
	tx           types.ITransaction
	txBody       types.ITransactionBody
	sender       hasharry.Address
	state        *types.ContractV2State
	events       []*types.Event
	height       uint64
	blockTime    uint64
}

func NewPairRunner(lib *library.RunnerLibrary, tx types.ITransaction, height, blockTime uint64) *PairRunner {
	var exchange *exchange2.Exchange
	var pair *exchange2.Pair
	var exchangeAddr string
	var addBody *exchange_func.ExchangeAddLiquidity
	var removeBody *exchange_func.ExchangeRemoveLiquidity
	txBody := tx.GetTxBody()
	contractBody, _ := txBody.(*types.TxContractV2Body)
	address := contractBody.Contract

	switch contractBody.FunctionType {
	case contractv2.Pair_AddLiquidity:
		addBody, _ = contractBody.Function.(*exchange_func.ExchangeAddLiquidity)
		exchangeAddr = addBody.Exchange.String()
	case contractv2.Pair_RemoveLiquidity:
		removeBody, _ = contractBody.Function.(*exchange_func.ExchangeRemoveLiquidity)
		exchangeAddr = removeBody.Exchange.String()
	}

	exHeader := lib.GetContractV2(exchangeAddr)
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
		addBody:      addBody,
		removeBody:   removeBody,
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
		events:       make([]*types.Event, 0),
	}
}

func (p *PairRunner) PreAddLiquidityVerify() error {
	if p.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
	}
	/*	if !p.sender.IsEqual(p.exchange.Admin) {
			return errors.New("forbidden")
		}
	*/
	if !p.addBody.TokenA.IsEqual(param.Token) {
		if contract := p.library.GetContract(p.addBody.TokenA.String()); contract == nil {
			return fmt.Errorf("tokenA %s is not exist", p.addBody.TokenA.String())
		}
	}
	if !p.addBody.TokenB.IsEqual(param.Token) {
		if contract := p.library.GetContract(p.addBody.TokenB.String()); contract == nil {
			return fmt.Errorf("tokenB %s is not exist", p.addBody.TokenB.String())
		}
	}

	address, err := PairAddress(param.Net, p.addBody.TokenA, p.addBody.TokenB, p.exHeader.Address)
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	if p.pair != nil {
		return p.preAddLiquidityVerify(address)
	}
	return nil
}

func (p *PairRunner) preAddLiquidityVerify(pairAddr string) error {
	if p.pair == nil {
		return errors.New("pair not exist")
	}
	pairContract := p.library.GetContractV2(pairAddr)
	if pairContract == nil {
		return fmt.Errorf("the pair %s is not exist", pairAddr)
	}
	pair := pairContract.Body.(*exchange2.Pair)
	reserveA, reserveB := p.library.GetReservesByPair(pair, p.addBody.TokenA, p.addBody.TokenB)
	amountA, amountB, err := p.optimalAmount(reserveA, reserveB, p.addBody.AmountADesired, p.addBody.AmountBDesired, p.addBody.AmountAMin, p.addBody.AmountBMin)
	if err != nil {
		return err
	}
	balanceA := p.library.GetBalance(p.sender, p.addBody.TokenA)
	if balanceA < amountA {
		return fmt.Errorf("insufficient balance %s", p.sender.String())
	}
	balanceB := p.library.GetBalance(p.sender, p.addBody.TokenB)
	if balanceB < amountB {
		return fmt.Errorf("insufficient balance %s", p.sender.String())
	}
	return nil
}

func (p *PairRunner) PreRemoveLiquidityVerify(lastHeight uint64) error {
	if p.removeBody.Deadline != 0 && p.removeBody.Deadline < lastHeight {
		return fmt.Errorf("past the deadline")
	}
	if p.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", p.removeBody.Exchange.String())
	}
	/*	if !p.sender.IsEqual(p.exchange.Admin) {
			return errors.New("forbidden")
		}
	*/
	if p.removeBody.Liquidity == 0 {
		return fmt.Errorf("invalid liquidity")
	}
	if !p.removeBody.TokenA.IsEqual(param.Token) {
		if contract := p.library.GetContract(p.removeBody.TokenA.String()); contract == nil {
			return fmt.Errorf("tokenA %s is not exist", p.removeBody.TokenA.String())
		}
	}
	if !p.removeBody.TokenB.IsEqual(param.Token) {
		if contract := p.library.GetContract(p.removeBody.TokenB.String()); contract == nil {
			return fmt.Errorf("tokenB %s is not exist", p.removeBody.TokenB.String())
		}
	}

	address, err := PairAddress(param.Net, p.removeBody.TokenA, p.removeBody.TokenB, p.exHeader.Address)
	if err != nil {
		return fmt.Errorf("pair address error")
	}
	if address != p.address.String() {
		return fmt.Errorf("wrong pair contract address")
	}
	if p.pair == nil {
		return fmt.Errorf("pair is not exist")
	}
	balance := p.library.GetBalance(p.sender, p.address)
	if balance < p.removeBody.Liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
	}
	token0, token1 := library.SortToken(p.removeBody.TokenA, p.removeBody.TokenB)
	_reserve0, _reserve1 := p.library.GetReservesByPair(p.pair, token0, token1)

	_liquidity := p.removeBody.Liquidity
	if balance < _liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
	}
	_totalSupply := p.pair.TotalSupply
	if _totalSupply < p.removeBody.Liquidity {
		return fmt.Errorf("%s's liquidity token is insufficient", p.address.String())
	}

	amount0 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve0))), big.NewInt(int64(_totalSupply))).Uint64()
	amount1 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve1))), big.NewInt(int64(_totalSupply))).Uint64()
	if token0.IsEqual(p.removeBody.TokenA) {
		if amount0 < p.removeBody.AmountAMin || amount1 < p.removeBody.AmountBMin {
			return fmt.Errorf("not meet expectations")
		}
	} else {
		if amount0 < p.removeBody.AmountBMin || amount1 < p.removeBody.AmountAMin {
			return fmt.Errorf("not meet expectations")
		}
	}
	return nil
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
				return 0, 0, errors.New("insufficient amountB")
			}
			return amountADesired, amountBOptimal, nil
		} else {
			// 则计算 最优数量A = 期望数量B * 储备A / 储备B
			amountAOptimal, err := p.quote(amountBDesired, reserveB, reserveA)
			if err != nil {
				return 0, 0, err
			}
			if amountAOptimal < amountAMin {
				return 0, 0, errors.New("insufficient amountA")
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
	if reserveA <= 0 || reserveB <= 0 {
		return 0, errors.New("insufficient_liquidity")
	}
	amountB := big.NewInt(0).Div(big.NewInt(0).Mul(big.NewInt(int64(amountA)), big.NewInt(int64(reserveB))), big.NewInt(int64(reserveA)))
	return amountB.Uint64(), nil
}

func (p *PairRunner) AddLiquidity() {
	var ERR error
	var err error
	var feeLiquidity uint64
	var feeOn bool
	defer func() {
		if ERR != nil {
			p.state.State = types.Contract_Failed
			p.state.Error = ERR.Error()
		} else {
			p.state.Event = p.events
		}
		p.state.Event = p.events
		p.library.SetContractV2State(p.tx.Hash().String(), p.state)
	}()
	if p.exHeader == nil {
		ERR = fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
		return
	}

	if p.pair == nil {
		p.createPair()
	}

	_reserveA, _reserveB := p.library.GetReservesByPair(p.pair, p.addBody.TokenA, p.addBody.TokenB)
	_reserve0, _reserve1 := p.pair.Reserve0, p.pair.Reserve1

	amountA, amountB, err := p.optimalAmount(_reserveA, _reserveB, p.addBody.AmountADesired, p.addBody.AmountBDesired, p.addBody.AmountAMin, p.addBody.AmountBMin)
	if err != nil {
		ERR = err
		return
	}

	liquidity, feeLiquidity, feeOn, err := p.mintLiquidityValue(_reserveA, _reserveB, amountA, amountB)
	if err != nil {
		ERR = err
		return
	}
	if p.addBody.TokenA.IsEqual(p.pair.Token0) {
		p.pair.UpdatePair(_reserve0+amountA, _reserve1+amountB, _reserve0, _reserve1, p.height, feeOn)
	} else {
		p.pair.UpdatePair(_reserve0+amountB, _reserve1+amountA, _reserve0, _reserve1, p.height, feeOn)
	}

	p.transferEvent(p.sender, p.address, p.addBody.TokenA, amountA)
	p.transferEvent(p.sender, p.address, p.addBody.TokenB, amountB)
	p.mintEvent(p.addBody.To, p.address, liquidity)
	if feeOn {
		p.mintEvent(p.exchange.FeeTo, p.address, feeLiquidity)
	}

	if err = p.runEvents(); err != nil {
		ERR = err
		return
	}
	p.update()
}

func (p *PairRunner) createPair() {
	token0, token1 := library.SortToken(p.addBody.TokenA, p.addBody.TokenB)
	symbol0, _ := p.library.ContractSymbol(token0)
	symbol1, _ := p.library.ContractSymbol(token1)
	p.pair = exchange2.NewPair(p.addBody.Exchange, token0, token1, symbol0, symbol1)

	p.pairHeader = &contractv2.ContractV2{
		Address:    p.address,
		CreateHash: p.tx.Hash(),
		Type:       contractv2.Pair_,
		Body:       p.pair,
	}
	p.exchange.AddPair(token0, token1, p.address)
}

type RemoveLiquidity struct {
	SymbolA string `json:"symbolA"`
	SymbolB string `json:"symbolB"`
	AmountA uint64 `json:"amountA"`
	AmountB uint64 `json:"amountB"`
}

func (p *PairRunner) RemoveLiquidity() {
	var ERR error
	defer func() {
		if ERR != nil {
			p.state.State = types.Contract_Failed
			p.state.Error = ERR.Error()
		} else {
			p.state.Event = p.events
		}
		p.library.SetContractV2State(p.tx.Hash().String(), p.state)
	}()

	if p.exHeader == nil {
		ERR = fmt.Errorf("exchange %s is not exist", p.addBody.Exchange.String())
		return
	}

	if p.pair == nil {
		ERR = errors.New("pair is not exist")
		return
	}

	token0, token1 := library.SortToken(p.removeBody.TokenA, p.removeBody.TokenB)
	_reserve0, _reserve1 := p.library.GetReservesByPair(p.pair, token0, token1)
	feeOn, feeLiquidity, err := p.mintFee(_reserve0, _reserve1)
	if err != nil {
		ERR = err
		return
	}
	balance := p.library.GetBalance(p.sender, p.address)
	_liquidity := p.removeBody.Liquidity
	if balance < _liquidity {
		ERR = fmt.Errorf("%s's liquidity token is insufficient", p.sender.String())
		return
	}
	_totalSupply := p.pair.TotalSupply
	if _totalSupply < p.removeBody.Liquidity {
		ERR = fmt.Errorf("%s's liquidity token is insufficient", p.address.String())
		return
	}

	amount0 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve0))), big.NewInt(int64(_totalSupply))).Uint64()
	amount1 := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(_liquidity)), big.NewInt(int64(_reserve1))), big.NewInt(int64(_totalSupply))).Uint64()
	if token0.IsEqual(p.removeBody.TokenA) {
		if amount0 < p.removeBody.AmountAMin || amount1 < p.removeBody.AmountBMin {
			ERR = fmt.Errorf("not meet expectations")
			return
		}
	} else {
		if amount0 < p.removeBody.AmountBMin || amount1 < p.removeBody.AmountAMin {
			ERR = fmt.Errorf("not meet expectations")
			return
		}
	}
	p.pair.UpdatePair(_reserve0-amount0, _reserve1-amount1, _reserve0, _reserve1, p.height, feeOn)

	if feeOn {
		p.mintEvent(p.exchange.FeeTo, p.address, feeLiquidity)
	}
	p.burnEvent(p.sender, p.address, p.removeBody.Liquidity)
	p.transferEvent(p.address, p.removeBody.To, token0, amount0)
	p.transferEvent(p.address, p.removeBody.To, token1, amount1)

	if err = p.runEvents(); err != nil {
		ERR = err
		return
	}
	p.update()
}

type mint struct {
	Address hasharry.Address
	Amount  uint64
}

func (p *PairRunner) mintLiquidityValue(_reserve0, _reserve1, amount0, amount1 uint64) (uint64, uint64, bool, error) {
	// must be defined here since totalSupply can update in mintFee
	_totalSupply := p.pair.TotalSupply
	// 返回铸造币的手续费开关
	feeOn, feeLiquidity, err := p.mintFee(_reserve0, _reserve1)
	if err != nil {
		return 0, 0, false, err
	}
	var liquidityValue uint64

	if _totalSupply == 0 {
		// sqrt(amount0 * amount1)
		liquidityBig := big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(amount1))))
		liquidityValue = liquidityBig.Uint64()
	} else {
		// valiquidityValue1 = amount0 / _reserve0 * _totalSupply
		// valiquidityValue2 = amount1 / _reserve1 * _totalSupply
		value0 := big.NewInt(0).Mul(big.NewInt(int64(amount0)), big.NewInt(int64(_totalSupply)))
		value1 := big.NewInt(0).Mul(big.NewInt(int64(amount1)), big.NewInt(int64(_totalSupply)))
		liquidityValue = math.Min(big.NewInt(0).Div(value0, big.NewInt(int64(_reserve0))).Uint64(), big.NewInt(0).Div(value1, big.NewInt(int64(_reserve1))).Uint64())
		if liquidityValue <= 0 {
			return 0, 0, false, errors.New("insufficient liquidity minted")
		}
	}

	return liquidityValue, feeLiquidity, feeOn, nil
}

// if fee is on, mint liquidity equivalent to 1/6th of the growth in sqrt(k)
func (p *PairRunner) mintFee(_reserve0, _reserve1 uint64) (bool, uint64, error) {
	var feeLiquidity uint64
	feeTo := p.exchange.FeeTo
	// 收费地址被设置，则收费开
	feeOn := !feeTo.IsEqual(hasharry.Address{})
	_kLast := p.pair.KLast // gas savings
	if feeOn {
		if _kLast.Cmp(big.NewInt(0)) != 0 {
			// rootK = Sqrt(_reserve0 * _reserve1)
			rootK := big.NewInt(0).Sqrt(big.NewInt(0).Mul(big.NewInt(int64(_reserve0)), big.NewInt(int64(_reserve1))))
			// rootKLast = Sqrt(_kLast)
			rootKLast := big.NewInt(0).Sqrt(_kLast)
			if rootK.Cmp(rootKLast) > 0 {
				// numerator = (rootK-rootKLast)*TotalSupply
				numerator := big.NewInt(0).Mul(big.NewInt(0).Sub(rootK, rootKLast), big.NewInt(int64(p.pair.TotalSupply)))
				// denominator =  * 5 + rootKLast
				denominator := big.NewInt(0).Add(big.NewInt(0).Mul(rootK, big.NewInt(5)), rootKLast)
				// liquidity = numerator / denominator
				liquidityBig := big.NewInt(0).Div(numerator, denominator)
				if liquidityBig.Cmp(big.NewInt(0)) > 0 {
					feeLiquidity = liquidityBig.Uint64()
				}
			}
		}
	} else if _kLast.Cmp(big.NewInt(0)) != 0 {
		p.pair.KLast = big.NewInt(0)
	}
	return feeOn, feeLiquidity, nil
}

func (p *PairRunner) update() {
	p.exHeader.Body = p.exchange
	p.pairHeader.Body = p.pair
	p.library.SetContractV2(p.exHeader)
	p.library.SetContractV2(p.pairHeader)
}

func (p *PairRunner) transferEvent(from, to, token hasharry.Address, amount uint64) {
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Transfer,
		From:      from,
		To:        to,
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) mintEvent(to, token hasharry.Address, amount uint64) {
	p.pair.Mint(amount)
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Mint,
		From:      hasharry.StringToAddress("mint"),
		To:        to,
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) burnEvent(from, token hasharry.Address, amount uint64) {
	p.pair.Burn(amount)
	p.events = append(p.events, &types.Event{
		EventType: types.Event_Burn,
		From:      from,
		To:        hasharry.StringToAddress("burn"),
		Token:     token,
		Amount:    amount,
		Height:    p.height,
	})
}

func (p *PairRunner) runEvents() error {
	for _, event := range p.events {
		if err := p.library.PreRunEvent(event); err != nil {
			return err
		}
	}
	for _, event := range p.events {
		p.library.RunEvent(event)
	}
	return nil
}

func PairAddress(net string, tokenA, tokenB hasharry.Address, exchange hasharry.Address) (string, error) {
	token0, token1 := library.SortToken(tokenA, tokenB)
	bytes := bytes2.Join([][]byte{[]byte(token0.String()), []byte(token1.String()), []byte(exchange.String())}, []byte{})
	return ut.GenerateContractV2Address(net, bytes)
}
