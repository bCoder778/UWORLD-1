package exchange_runner

import (
	"fmt"
	"github.com/uworldao/UWORLD/common/codec"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
	"github.com/uworldao/UWORLD/ut"
)

type ExchangeRunner struct {
	library *library.RunnerLibrary
}

func NewExchangeRunner(lib *library.RunnerLibrary) *ExchangeRunner {
	return &ExchangeRunner{library: lib}
}

func (e *ExchangeRunner) PreVerify(from hasharry.Address, contract hasharry.Address, funcType contractv2.FunctionType) error {
	conV2 := e.library.GetContractV2(contract.String())
	switch funcType {
	case contractv2.Exchange_Init_:
		if conV2 != nil {
			return fmt.Errorf("exchange %s already exist", contract.String())
		}
	case contractv2.Exchange_SetAdmin_:
		fallthrough
	case contractv2.Exchange_SetFeeTo_:
		if conV2 == nil {
			return fmt.Errorf("exchanges %s is not exist", contract.String())
		}
		ex := conV2.Body.(*exchange.Exchange)
		return ex.VerifySetter(from)
	}
	return nil
}

func (e *ExchangeRunner) Init(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	contract := &contractv2.ContractV2{
		Address:    body.GetContract(),
		CreateHash: head.TxHash,
		Type:       body.Type,
		Body:       nil,
	}
	contractV2 := e.library.GetContractV2(contract.Address.String())
	if contractV2 != nil {
		body.State = types.Contract_Failed
		body.Message = fmt.Sprintf("exchange %s already exist", contract.Address.String())
		return
	}
	initBody := body.Function.(*exchange_func.ExchangeInitBody)
	contract.Body = exchange.NewExchange(initBody.Admin, initBody.FeeTo)
	e.library.SetContractV2(contract)
	body.State = types.Contract_Success
}

func (e *ExchangeRunner) SetAdmin(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	ctrV2 := e.library.GetContractV2(body.Contract.String())
	if ctrV2 == nil {
		body.State = types.Contract_Failed
		body.Message = fmt.Sprintf("exchanges %s is not exist", body.Contract.String())
		return
	}
	funcBody, _ := body.Function.(*exchange_func.ExchangeAdmin)
	ex, _ := ctrV2.Body.(*exchange.Exchange)
	if err := ex.SetAdmin(funcBody.Address, head.From); err != nil {
		body.State = types.Contract_Failed
		body.Message = err.Error()
	}
	ctrV2.Body = ex
	e.library.SetContractV2(ctrV2)
	body.State = types.Contract_Success
}

func (e *ExchangeRunner) SetFeeTo(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	ctrV2 := e.library.GetContractV2(body.Contract.String())
	if ctrV2 == nil {
		body.State = types.Contract_Failed
		body.Message = fmt.Sprintf("exchanges %s is not exist", body.Contract.String())
	}
	funcBody, _ := body.Function.(*exchange_func.ExchangeFeeTo)
	ex, _ := ctrV2.Body.(*exchange.Exchange)
	if err := ex.SetFeeTo(funcBody.Address, head.From); err != nil {
		body.State = types.Contract_Failed
		body.Message = err.Error()
	}
	ctrV2.Body = ex
	e.library.SetContractV2(ctrV2)
	body.State = types.Contract_Success
}

func ExchangeAddress(net, from string, nonce uint64) (string, error) {
	bytes := make([]byte, 0)
	nonceBytes := codec.Uint64toBytes(nonce)
	bytes = append(hasharry.StringToAddress(from).Bytes(), nonceBytes...)
	return ut.GenerateContractV2Address(net, bytes)
}
