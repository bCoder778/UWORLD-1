package exchange_runner

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/codec"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
	"math/big"
)

type ExchangeRunner struct {
	library      *library.RunnerLibrary
	exHeader     *contractv2.ContractV2
	exchange     *exchange.Exchange
	address      hasharry.Address
	tx           types.ITransaction
	contractBody *types.TxContractV2Body
	transList    []*library.TransferInfo
	pairList     []*contractv2.ContractV2
}

func NewExchangeRunner(lib *library.RunnerLibrary, tx types.ITransaction) *ExchangeRunner {
	var ex *exchange.Exchange
	address := tx.GetTxBody().GetContract()
	exHeader := lib.GetContractV2(address.String())
	if exHeader != nil {
		ex = exHeader.Body.(*exchange.Exchange)
	}

	contractBody := tx.GetTxBody().(*types.TxContractV2Body)
	return &ExchangeRunner{library: lib,
		exHeader:     exHeader,
		address:      address,
		tx:           tx,
		exchange:     ex,
		contractBody: contractBody,
		transList:    make([]*library.TransferInfo, 0),
	}
}

func (e *ExchangeRunner) PreInitVerify() error {
	if e.exHeader != nil {
		return fmt.Errorf("exchange %s already exist", e.address.String())
	}
	return nil
}

func (e *ExchangeRunner) PreSetVerify() error {
	if e.exHeader == nil {
		return fmt.Errorf("exchange %s is not exist", e.address.String())
	}
	return e.exchange.VerifySetter(e.tx.From())
}

func (e *ExchangeRunner) PreExactInVerify(lastHeight uint64) error {
	if e.exHeader == nil {
		return fmt.Errorf("exchange is not exist")
	}
	funcBody, _ := e.contractBody.Function.(*exchange_func.ExactIn)
	if funcBody == nil {
		return errors.New("wrong contractV2 function")
	}
	if len(funcBody.Path) < 2 {
		return errors.New("invalid path")
	}
	for i := 0; i < len(funcBody.Path)-1; i++ {
		if exist := e.exchange.Exist(library.SortToken(funcBody.Path[i], funcBody.Path[i+1])); !exist {
			return fmt.Errorf("the pair of %s and %s does not exist", funcBody.Path[i].String(), funcBody.Path[i+1].String())
		}
	}
	if funcBody.Deadline != 0 && funcBody.Deadline < lastHeight {
		return fmt.Errorf("past the deadline")
	}
	balance := e.library.GetBalance(e.tx.From(), funcBody.Path[0])
	if funcBody.Path[0].IsEqual(param.Token) {
		if balance < funcBody.AmountIn+e.tx.GetFees() {
			return errors.New("balance not enough")
		}
	} else {
		if balance < funcBody.AmountIn {
			return errors.New("balance not enough")
		}
	}
	return nil
}

func (e *ExchangeRunner) PreExactOutVerify(lastHeight uint64) error {
	if e.exHeader == nil {
		return fmt.Errorf("exchange is not exist")
	}
	funcBody, _ := e.contractBody.Function.(*exchange_func.ExactOut)
	if funcBody == nil {
		return errors.New("wrong contractV2 function")
	}
	if len(funcBody.Path) < 2 {
		return errors.New("invalid path")
	}
	for i := 0; i < len(funcBody.Path)-1; i++ {
		if exist := e.exchange.Exist(library.SortToken(funcBody.Path[i], funcBody.Path[i+1])); !exist {
			return fmt.Errorf("the pair of %s and %s does not exist", funcBody.Path[i].String(), funcBody.Path[i+1].String())
		}
	}
	if funcBody.Deadline != 0 && funcBody.Deadline < lastHeight {
		return fmt.Errorf("past the deadline")
	}
	balance := e.library.GetBalance(e.tx.From(), funcBody.Path[0])
	if funcBody.Path[0].IsEqual(param.Token) {
		if balance < funcBody.AmountInMax+e.tx.GetFees() {
			return errors.New("balance not enough")
		}
	} else {
		if balance < funcBody.AmountInMax {
			return errors.New("balance not enough")
		}
	}
	return nil
}

func (e *ExchangeRunner) Init() {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(e.tx.Hash().String(), state)
	}()

	contract := &contractv2.ContractV2{
		Address:    e.contractBody.Contract,
		CreateHash: e.tx.Hash(),
		Type:       e.contractBody.Type,
		Body:       nil,
	}
	if e.exHeader != nil {
		ERR = fmt.Errorf("exchange %s already exist", contract.Address.String())
		return
	}
	initBody := e.contractBody.Function.(*exchange_func.ExchangeInitBody)
	contract.Body = exchange.NewExchange(initBody.Admin, initBody.FeeTo)
	e.library.SetContractV2(contract)
}

func (e *ExchangeRunner) SetAdmin() {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(e.tx.Hash().String(), state)
	}()

	if e.exHeader == nil {
		ERR = fmt.Errorf("exchanges %s is not exist", e.tx.GetTxBody().GetContract().String())
		return
	}
	funcBody, _ := e.contractBody.Function.(*exchange_func.ExchangeAdmin)
	ex, _ := e.exHeader.Body.(*exchange.Exchange)
	if err := ex.SetAdmin(funcBody.Address, e.tx.From()); err != nil {
		ERR = err
		return
	}
	e.exHeader.Body = ex
	e.library.SetContractV2(e.exHeader)
}

func (e *ExchangeRunner) SetFeeTo() {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(e.tx.Hash().String(), state)
	}()

	if e.exHeader == nil {
		ERR = fmt.Errorf("exchanges %s is not exist", e.tx.GetTxBody().GetContract().String())
		return
	}
	funcBody, _ := e.contractBody.Function.(*exchange_func.ExchangeFeeTo)
	ex, _ := e.exHeader.Body.(*exchange.Exchange)
	if err := ex.SetFeeTo(funcBody.Address, e.tx.From()); err != nil {
		ERR = err
		return
	}
	e.exHeader.Body = ex
	e.library.SetContractV2(e.exHeader)
}

func (e *ExchangeRunner) SwapExactIn(blockHeight, blockTime uint64) {
	var ERR error
	var err error
	var amounts []uint64
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		} else {
			state.Message = swapInfo(amounts)
		}
		e.library.SetContractV2State(e.tx.Hash().String(), state)
	}()

	funcBody, _ := e.contractBody.Function.(*exchange_func.ExactIn)

	if funcBody.Deadline != 0 && funcBody.Deadline < blockHeight {
		ERR = fmt.Errorf("past the deadline")
		return
	}
	amounts, err = e.getAmountsOut(funcBody.AmountIn, funcBody.Path)
	if err != nil {
		ERR = err
		return
	}
	outAmount := amounts[len(amounts)-1]
	if outAmount < funcBody.AmountOutMin {
		ERR = fmt.Errorf("outAmount %d is less than the minimum output %d", outAmount, funcBody.AmountOutMin)
		return
	}
	pair0 := e.exchange.PairAddress(library.SortToken(funcBody.Path[0], funcBody.Path[1]))
	if err := e.swapAmounts(amounts, funcBody.Path, funcBody.To, blockHeight, blockTime); err != nil {
		ERR = err
		return
	}
	transferInfo := &library.TransferInfo{
		From:   e.tx.From(),
		To:     pair0,
		Token:  funcBody.Path[0],
		Amount: amounts[0],
		Height: blockHeight,
	}
	if err := e.library.PreTransfer(transferInfo); err != nil {
		ERR = err
		return
	}
	e.transList = append(e.transList, transferInfo)
	e.transfer()
	e.update()
}

func (e *ExchangeRunner) transfer() {
	for _, info := range e.transList {
		e.library.Transfer(info)
	}
}

func (e *ExchangeRunner) update() {
	for _, pairContract := range e.pairList {
		e.library.SetContractV2(pairContract)
	}
}

// requires the initial amount to have already been sent to the first pair
func (e *ExchangeRunner) swapAmounts(amounts []uint64, path []hasharry.Address, to hasharry.Address, height, blockTime uint64) error {
	var amount0Out, amount1Out uint64
	var amount0In, amount1In uint64
	for i := 0; i < len(path)-1; i++ {
		input, output := path[i], path[i+1]
		token0, _ := library.SortToken(input, output)
		amountOut := amounts[i+1]
		amountIn := amounts[i]
		if input.IsEqual(token0) {
			amount0Out, amount1Out = 0, amountOut
			amount0In, amount1In = amountIn, 0
		} else {
			amount0Out, amount1Out = amountOut, 0
			amount0In, amount1In = 0, amountIn
		}
		toAddr := to
		if i < len(path)-2 {
			toAddr = e.exchange.PairAddress(library.SortToken(output, path[i+2]))
		}
		if err := e.swap(input, output, amount0In, amount1In, amount0Out, amount1Out, toAddr, height, blockTime); err != nil {
			return err
		}
	}
	return nil
}

func (e *ExchangeRunner) swap(tokenA, tokenB hasharry.Address, amount0In, amount1In, amount0Out, amount1Out uint64, to hasharry.Address, height uint64, blockTime uint64) error {
	if amount0Out <= 0 && amount1Out <= 0 {
		return errors.New("insufficient output amount")
	}
	_token0, _token1 := library.SortToken(tokenA, tokenB)
	pairAddress := e.exchange.PairAddress(_token0, _token1)
	pairContract := e.library.GetContractV2(pairAddress.String())
	pair := pairContract.Body.(*exchange.Pair)
	_reserve0, _reserve1 := e.library.GetReservesByPairAddress(pairAddress, _token0, _token1)
	if amount0Out >= _reserve0 || amount1Out >= _reserve1 {
		return errors.New("insufficient liquidity")
	}

	var balance0, balance1 uint64
	if to.IsEqual(_token0) || to.IsEqual(_token1) {
		return errors.New("invalid to")
	}
	// 转账给to地址
	if amount0Out > 0 {
		transInfo := &library.TransferInfo{
			From:   pairAddress,
			To:     to,
			Token:  _token0,
			Amount: amount0Out,
			Height: height,
		}
		if err := e.library.PreTransfer(transInfo); err != nil {
			return err
		}
		e.transList = append(e.transList, transInfo)
	}
	if amount1Out > 0 {
		transInfo := &library.TransferInfo{
			From:   pairAddress,
			To:     to,
			Token:  _token1,
			Amount: amount1Out,
			Height: height,
		}
		if err := e.library.PreTransfer(transInfo); err != nil {
			return err
		}
		e.transList = append(e.transList, transInfo)
	}

	balance0 = e.library.GetBalance(pairAddress, _token0)
	balance1 = e.library.GetBalance(pairAddress, _token1)
	if amount0In > 0 {
		balance0 = balance0 + amount0In
	} else {
		balance1 = balance1 + amount1In
	}

	if amount0Out > 0 {
		if balance0 < amount0Out {
			return errors.New("insufficient liquidity")
		}
		balance0 = balance0 - amount0Out
	} else {
		if balance1 < amount1Out {
			return errors.New("insufficient liquidity")
		}
		balance1 = balance1 - amount1Out
	}

	//通过输出数量，算输入数量
	if balance0 > _reserve0-amount0Out {
		amount0In = balance0 - (_reserve0 - amount0Out)
	} else {
		amount0In = 0
	}

	if balance1 > _reserve1-amount1Out {
		amount1In = balance1 - (_reserve1 - amount1Out)
	} else {
		amount1In = 0
	}
	if amount0In <= 0 && amount1In <= 0 {
		return errors.New("insufficient input amount")
	}
	// balance0Adjusted = balance0 * 1000 - amount0In * 3
	balance0Adjusted := big.NewInt(0).Sub(big.NewInt(0).Mul(big.NewInt(int64(balance0)), big.NewInt(1000)),
		big.NewInt(0).Mul(big.NewInt(int64(amount0In)), big.NewInt(3)))
	// balance1Adjusted = balance1 * 1000 - amount1In * 3
	balance1Adjusted := big.NewInt(0).Sub(big.NewInt(0).Mul(big.NewInt(int64(balance1)), big.NewInt(1000)),
		big.NewInt(0).Mul(big.NewInt(int64(amount1In)), big.NewInt(3)))

	// 确保k值大于K值，判断是否已经收过税
	// x = balance0Adjusted * balance1Adjusted
	x := big.NewInt(0).Mul(balance0Adjusted, balance1Adjusted)
	// y = _reserve0 * _reserve1 * 1000^2
	y := big.NewInt(0).Mul(big.NewInt(0).Mul(big.NewInt(int64(_reserve0)), big.NewInt(int64(_reserve1))), big.NewInt(1000^2))
	if x.Cmp(y) < 0 {
		return errors.New("K")
	}
	pair.Update(balance0, balance1, _reserve0, _reserve1, blockTime)
	pairContract.Body = pair
	e.pairList = append(e.pairList, pairContract)
	return nil
}

// performs chained getAmountOut calculations on any number of pairs
func (e *ExchangeRunner) getAmountsOut(amountIn uint64, path []hasharry.Address) ([]uint64, error) {
	var err error
	amounts := make([]uint64, len(path))
	amounts[0] = amountIn
	for i := 0; i < len(path)-1; i++ {
		// 获取储备量
		token0, token1 := library.SortToken(path[i], path[i+1])
		pairAddress := e.exchange.PairAddress(token0, token1)
		reserveIn, reserveOut := e.library.GetReservesByPairAddress(pairAddress, path[i], path[i+1])
		// 下一个数额 =  当前数额兑换的结果
		amounts[i+1], err = GetAmountOut(amounts[i], reserveIn, reserveOut)
		if err != nil {
			return amounts, err
		}
	}
	return amounts, nil
}

// getAmountsIn performs chained getAmountIn calculations on any number of pairs
func (e *ExchangeRunner) getAmountsIn(amountOut uint64, path []hasharry.Address) ([]uint64, error) {
	var err error
	amounts := make([]uint64, len(path))
	amounts[len(amounts)-1] = amountOut
	for i := len(path) - 1; i > 0; i-- {
		// 获取储备量
		token0, token1 := library.SortToken(path[i-1], path[i])
		pairAddress := e.exchange.PairAddress(token0, token1)
		reserveIn, reserveOut := e.library.GetReservesByPairAddress(pairAddress, path[i-1], path[i])
		amounts[i-1], err = GetAmountIn(amounts[i], reserveIn, reserveOut)
		if err != nil {
			return amounts, err
		}
	}
	return amounts, nil
}

// GetAmountOut given an input amount of an asset and pair reserves, returns the maximum output amount of the other asset
func GetAmountOut(amountIn, reserveIn, reserveOut uint64) (uint64, error) {
	if amountIn <= 0 {
		return 0, errors.New("insufficient input amount")
	}
	if reserveIn <= 0 || reserveOut <= 0 {
		return 0, errors.New("insufficient liquidity")
	}
	// amountInWithFee = amountIn * 995
	// 0.5% fees
	amountInWithFee := big.NewInt(0).Mul(big.NewInt(int64(amountIn)), big.NewInt(995))
	// numerator = amountInWithFee * reserveOut
	numerator := big.NewInt(0).Mul(amountInWithFee, big.NewInt(int64(reserveOut)))
	// denominator = reserveIn * 1000 + amountInWithFee

	denominator := big.NewInt(0).Add(big.NewInt(0).Mul(big.NewInt(int64(reserveIn)), big.NewInt(1000)), amountInWithFee)
	amountOut := big.NewInt(0).Div(numerator, denominator)
	return amountOut.Uint64(), nil
}

// GetAmountIn given an output amount of an asset and pair reserves, returns a required input amount of the other asset
func GetAmountIn(amountOut, reserveIn, reserveOut uint64) (uint64, error) {
	if amountOut <= 0 {
		return 0, errors.New("insufficient output amount")
	}
	if reserveIn <= 0 || reserveOut <= 0 {
		return 0, errors.New("insufficient liquidity")
	}
	if reserveOut < amountOut {
		return 0, errors.New("insufficient liquidity")
	}
	/*	amountOut = amountOut / 1000000
		reserveIn = reserveIn / 10000000
		reserveOut = reserveOut / 10000000*/
	// numerator = amountOut * reserveIn * 1000
	numerator := big.NewInt(0).Mul(big.NewInt(0).Mul(big.NewInt(int64(amountOut)), big.NewInt(int64(reserveIn))), big.NewInt(1000))
	// denominator = (reserveOut - amountOut) (* 995)
	denominator := big.NewInt(0).Mul(big.NewInt(0).Sub(big.NewInt(int64(reserveOut)), big.NewInt(int64(amountOut))), big.NewInt(995))
	// amountIn = (numerator\denominator) + 1
	x := big.NewInt(0).Div(numerator, denominator)

	amountIn := big.NewInt(0).Add(x, big.NewInt(1))
	return amountIn.Uint64(), nil
}

func (e *ExchangeRunner) SwapExactOut(blockHeight uint64, blockTime uint64) {
	var ERR error
	var err error
	var amounts []uint64
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		} else {
			state.Message = swapInfo(amounts)
		}
		e.library.SetContractV2State(e.tx.Hash().String(), state)
	}()

	funcBody, _ := e.contractBody.Function.(*exchange_func.ExactOut)

	if funcBody.Deadline != 0 && funcBody.Deadline < blockHeight {
		ERR = fmt.Errorf("past the deadline")
		return
	}
	amounts, err = e.getAmountsIn(funcBody.AmountOut, funcBody.Path)
	if err != nil {
		ERR = err
		return
	}
	if amounts[0] > funcBody.AmountInMax {
		ERR = fmt.Errorf("amountIn %d is greater than the maximum input amount %d", amounts[0], funcBody.AmountInMax)
		return
	}
	pair0 := e.exchange.PairAddress(library.SortToken(funcBody.Path[0], funcBody.Path[1]))
	if err := e.swapAmounts(amounts, funcBody.Path, funcBody.To, blockHeight, blockTime); err != nil {
		ERR = err
		return
	}
	transferInfo := &library.TransferInfo{
		From:   e.tx.From(),
		To:     pair0,
		Token:  funcBody.Path[0],
		Amount: amounts[0],
		Height: blockHeight,
	}
	if err := e.library.PreTransfer(transferInfo); err != nil {
		ERR = err
		return
	}
	e.transList = append(e.transList, transferInfo)
	e.transfer()
	e.update()
}

func ExchangeAddress(net, from string, nonce uint64) (string, error) {
	bytes := make([]byte, 0)
	nonceBytes := codec.Uint64toBytes(nonce)
	bytes = append([]byte(from), nonceBytes...)
	return ut.GenerateContractV2Address(net, bytes)
}

func swapInfo(amounts []uint64) string {
	str := ""
	for i, amopunt := range amounts {
		if i != len(amounts)-1 {
			str += fmt.Sprintf("%d-", amopunt)
		} else {
			str += fmt.Sprintf("%d", amopunt)
		}
	}
	return str
}
