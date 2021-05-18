package contractstate

import (
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody"
	"github.com/uworldao/UWORLD/database/contractdb"
	"sync"
)

const contractSate = "contract_state"

// Contract status, used to store all published contract information records
type ContractState struct {
	contractDb      IContractStorage
	contractMutex   sync.RWMutex
	confirmedHeight uint64
}

func NewContractState(dataDir string) (*ContractState, error) {
	storage := contractdb.NewContractStorage(dataDir + "/" + contractSate)
	err := storage.Open()
	if err != nil {
		return nil, err
	}
	return &ContractState{
		contractDb: storage,
	}, nil
}

// Initialize the contract state tree
func (cs *ContractState) InitTrie(contractRoot hasharry.Hash) error {
	return cs.contractDb.InitTrie(contractRoot)
}

func (cs *ContractState) RootHash() hasharry.Hash {
	return cs.contractDb.RootHash()
}

// Commit contract status changes
func (cs *ContractState) ContractTrieCommit() (hasharry.Hash, error) {
	return cs.contractDb.Commit()
}

func (c *ContractState) GetContract(contractAddr string) *types.Contract {
	c.contractMutex.RLock()
	defer c.contractMutex.RUnlock()

	contract := c.contractDb.GetContractState(contractAddr)
	return contract
}

func (c *ContractState) UpdateConfirmedHeight(height uint64) {
	c.confirmedHeight = height
}

// Verification contract
func (c *ContractState) VerifyState(tx types.ITransaction) error {
	c.contractMutex.RLock()
	defer c.contractMutex.RUnlock()

	switch tx.GetTxType() {
	case types.Contract_:
		contractAddr := tx.GetTxBody().GetContract()
		contract := c.contractDb.GetContractState(contractAddr.String())
		if contract != nil {
			return contract.Verify(tx)
		}
	case types.ContractV2_:
		body, _ := tx.GetTxBody().(*types.ContractV2Body)
		contractAddr := tx.GetTxBody().GetContract()
		contract := c.contractDb.GetContractV2State(contractAddr.String())
		if contract != nil {
			return contract.Verify(body.Type, body.FunctionType, tx.From())
		}
	}

	return nil
}

func (c *ContractState) UpdateContract(tx types.ITransaction, blockHeight uint64) {
	c.contractMutex.Lock()
	defer c.contractMutex.Unlock()

	if tx.GetTxType() != types.Contract_ {
		return
	}
	txBody := tx.GetTxBody()
	contractRecord := &types.ContractRecord{
		Height:   blockHeight,
		TxHash:   tx.Hash(),
		Time:     tx.GetTime(),
		Amount:   txBody.GetAmount(),
		Receiver: txBody.ToAddress().String(),
	}
	contractAddr := txBody.GetContract()
	contract := c.contractDb.GetContractState(contractAddr.String())
	if contract != nil {
		contract.AddContract(contractRecord)
	} else {
		contract = &types.Contract{
			Contract:       contractAddr.String(),
			CoinName:       txBody.GetName(),
			CoinAbbr:       txBody.GetAbbr(),
			Description:    txBody.GetDescription(),
			IncreaseSwitch: txBody.GetIncreaseSwitch(),
			Records: &types.RecordList{
				contractRecord,
			},
		}
	}
	c.contractDb.SetContractState(contract)
}

func (c *ContractState) UpdateContractV2(tx types.ITransaction, blockHeight uint64) error {
	c.contractMutex.Lock()
	defer c.contractMutex.Unlock()

	body, _ := tx.GetTxBody().(*types.ContractV2Body)
	switch body.FunctionType {
	case contractv2.Exchange_Init_:
		return c.exchangeInit(tx.GetTxHead(), body, blockHeight)
	case contractv2.Exchange_SetFeeToSetter_:
		return c.exchangeSetFeeToSetter(tx.GetTxHead(), body, blockHeight)
	case contractv2.Exchange_SetFeeTo_:
		return c.exchangeSetFeeTo(tx.GetTxHead(), body, blockHeight)
	}
	return nil
}

func (c *ContractState) exchangeInit(head *types.TransactionHead, body *types.ContractV2Body, height uint64) error {
	contract := &contractv2.ContractV2{
		Address:    body.GetContract(),
		CreateHash: head.TxHash,
		Type:       body.Type,
		Body:       nil,
	}
	contractV2 := c.contractDb.GetContractV2State(contract.Address.String())
	if contractV2 != nil {
		return fmt.Errorf("exchanges %s already exist", contract.Address.String())
	}
	initBody := body.Function.(*functionbody.ExchangeInitBody)
	contract.Body = contractv2.NewExchange(initBody.FeeToSetter, initBody.FeeTo)
	c.contractDb.SetContractV2State(contract)
	return nil
}

func (c *ContractState) exchangeSetFeeToSetter(head *types.TransactionHead, body *types.ContractV2Body, height uint64) error {
	ctrV2 := c.contractDb.GetContractV2State(body.Contract.String())
	if ctrV2 == nil {
		return fmt.Errorf("exchanges %s is not exist", body.Contract.String())
	}
	funcBody, _ := body.Function.(*functionbody.ExchangeFeeToSetter)
	ex, _ := ctrV2.Body.(*contractv2.Exchange)
	if err := ex.SetFeeToSetter(funcBody.Address, head.From); err != nil {
		return err
	}
	ctrV2.Body = ex
	c.contractDb.SetContractV2State(ctrV2)
	return nil
}

func (c *ContractState) exchangeSetFeeTo(head *types.TransactionHead, body *types.ContractV2Body, height uint64) error {
	ctrV2 := c.contractDb.GetContractV2State(body.Contract.String())
	if ctrV2 == nil {
		return fmt.Errorf("exchanges %s is not exist", body.Contract.String())
	}
	funcBody, _ := body.Function.(*functionbody.ExchangeFeeTo)
	ex, _ := ctrV2.Body.(*contractv2.Exchange)
	if err := ex.SetFeeTo(funcBody.Address, head.From); err != nil {
		return err
	}
	ctrV2.Body = ex
	c.contractDb.SetContractV2State(ctrV2)
	return nil
}

func (c *ContractState) Close() error {
	return c.contractDb.Close()
}
