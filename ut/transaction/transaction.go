package transaction

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
	"time"
)

func NewTransaction(from, to, token string, note string, amount, nonce uint64) *types.Transaction {
	tx := &types.Transaction{
		TxHead: &types.TransactionHead{
			TxType:     types.Transfer_,
			TxHash:     hasharry.Hash{},
			From:       hasharry.StringToAddress(from),
			Nonce:      nonce,
			Time:       uint64(time.Now().Unix()),
			Note:       note,
			SignScript: &types.SignScript{},
			Fees:       param.Fees,
		},
		TxBody: &types.NormalTransactionBody{
			Contract: hasharry.StringToAddress(token),
			To:       hasharry.StringToAddress(to),
			Amount:   amount,
		},
	}
	tx.SetHash()
	return tx
}

func NewContract(from, to, contract string, note string, amount, nonce uint64, name, abbr string, increase bool, description string) *types.Transaction {
	tx := &types.Transaction{
		TxHead: &types.TransactionHead{
			TxType:     types.Contract_,
			TxHash:     hasharry.Hash{},
			From:       hasharry.StringToAddress(from),
			Nonce:      nonce,
			Time:       uint64(time.Now().Unix()),
			Note:       note,
			SignScript: &types.SignScript{},
			Fees:       param.TokenConsumption,
		},
		TxBody: &types.ContractBody{
			Contract:       hasharry.StringToAddress(contract),
			To:             hasharry.StringToAddress(to),
			Name:           name,
			Abbr:           abbr,
			IncreaseSwitch: increase,
			Description:    description,
			Amount:         amount,
		},
	}
	tx.SetHash()
	return tx
}

func NewExchange(net, from, feeToSetter, feeTo string, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := ut.GenerateContractV2Address(net, from, nonce)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{
		TxHead: &types.TransactionHead{
			TxType:     types.ContractV2_,
			TxHash:     hasharry.Hash{},
			From:       hasharry.StringToAddress(from),
			Nonce:      nonce,
			Time:       uint64(time.Now().Unix()),
			Note:       note,
			SignScript: &types.SignScript{},
			Fees:       param.Fees,
		},
		TxBody: &types.ContractV2Body{
			Contract:     hasharry.StringToAddress(contract),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_Init_,
			Function: &functionbody.ExchangeInitBody{
				FeeToSetter: hasharry.StringToAddress(feeToSetter),
				FeeTo:       hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}
