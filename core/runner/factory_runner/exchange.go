package factory_runner

import (
	"fmt"
	"github.com/uworldao/UWORLD/common/codec"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/runner/library"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/contractv2/factory"
	"github.com/uworldao/UWORLD/core/types/functionbody/factory_func"
	"github.com/uworldao/UWORLD/ut"
)

type FactoryRunner struct {
	library *library.RunnerLibrary
}

func NewFactoryRunner(lib *library.RunnerLibrary) *FactoryRunner {
	return &FactoryRunner{library: lib}
}

func (e *FactoryRunner) PreVerify(from hasharry.Address, contract hasharry.Address, funcType contractv2.FunctionType) error {
	conV2 := e.library.GetContractV2(contract.String())
	switch funcType {
	case contractv2.Factory_Init_:
		if conV2 != nil {
			return fmt.Errorf("factory %s already exist", contract.String())
		}
	case contractv2.Factory_SetAdmin_:
		fallthrough
	case contractv2.Factory_SetFeeTo_:
		if conV2 == nil {
			return fmt.Errorf("factorys %s is not exist", contract.String())
		}
		ex := conV2.Body.(*factory.Factory)
		return ex.VerifySetter(from)
	}
	return nil
}

func (e *FactoryRunner) Init(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(head.TxHash.String(), state)

	}()

	contract := &contractv2.ContractV2{
		Address:    body.GetContract(),
		CreateHash: head.TxHash,
		Type:       body.Type,
		Body:       nil,
	}
	contractV2 := e.library.GetContractV2(contract.Address.String())
	if contractV2 != nil {
		ERR = fmt.Errorf("factory %s already exist", contract.Address.String())
		return
	}
	initBody := body.Function.(*factory_func.FactoryInitBody)
	contract.Body = factory.NewFactory(initBody.Admin, initBody.FeeTo)
	e.library.SetContractV2(contract)
}

func (e *FactoryRunner) SetAdmin(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(head.TxHash.String(), state)
	}()

	ctrV2 := e.library.GetContractV2(body.Contract.String())
	if ctrV2 == nil {
		ERR = fmt.Errorf("factorys %s is not exist", body.Contract.String())
		return
	}
	funcBody, _ := body.Function.(*factory_func.FactoryAdmin)
	ex, _ := ctrV2.Body.(*factory.Factory)
	if err := ex.SetAdmin(funcBody.Address, head.From); err != nil {
		ERR = err
		return
	}
	ctrV2.Body = ex
	e.library.SetContractV2(ctrV2)
}

func (e *FactoryRunner) SetFeeTo(head *types.TransactionHead, body *types.ContractV2Body, height uint64) {
	var ERR error
	state := &types.ContractV2State{State: types.Contract_Success}
	defer func() {
		if ERR != nil {
			state.State = types.Contract_Failed
			state.Message = ERR.Error()
		}
		e.library.SetContractV2State(head.TxHash.String(), state)
	}()

	ctrV2 := e.library.GetContractV2(body.Contract.String())
	if ctrV2 == nil {
		ERR = fmt.Errorf("factorys %s is not exist", body.Contract.String())
		return
	}
	funcBody, _ := body.Function.(*factory_func.FactoryFeeTo)
	ex, _ := ctrV2.Body.(*factory.Factory)
	if err := ex.SetFeeTo(funcBody.Address, head.From); err != nil {
		ERR = err
		return
	}
	ctrV2.Body = ex
	e.library.SetContractV2(ctrV2)
}

func FactoryAddress(net, from string, nonce uint64) (string, error) {
	bytes := make([]byte, 0)
	nonceBytes := codec.Uint64toBytes(nonce)
	bytes = append(hasharry.StringToAddress(from).Bytes(), nonceBytes...)
	return ut.GenerateContractV2Address(net, bytes)
}
