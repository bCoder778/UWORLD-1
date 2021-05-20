package transaction

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/runner/factory_runner"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody/factory_func"
	"github.com/uworldao/UWORLD/param"
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

func NewFactory(net, from, admin, feeTo string, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := factory_runner.FactoryAddress(net, from, nonce)
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
			Type:         contractv2.Factory_,
			FunctionType: contractv2.Factory_Init_,
			Function: &factory_func.FactoryInitBody{
				Admin: hasharry.StringToAddress(admin),
				FeeTo: hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSetAdmin(from, factory, admin string, nonce uint64, note string) (*types.Transaction, error) {
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
			Contract:     hasharry.StringToAddress(factory),
			Type:         contractv2.Factory_,
			FunctionType: contractv2.Factory_SetAdmin_,
			Function: &factory_func.FactoryAdmin{
				Address: hasharry.StringToAddress(admin),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSetFeeTo(from, factory, feeTo string, nonce uint64, note string) (*types.Transaction, error) {
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
			Contract:     hasharry.StringToAddress(factory),
			Type:         contractv2.Factory_,
			FunctionType: contractv2.Factory_SetFeeTo_,
			Function: &factory_func.FactoryFeeTo{
				Address: hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewPairCreate(net, from, to, factory, tokenA, tokenB string, amountADesired, amountBDesired, amountAMin, amountBMin, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := factory_runner.PairAddress(net, tokenA, tokenB)
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
			Type:         contractv2.Pair_,
			FunctionType: contractv2.Pair_Create,
			Function: &factory_func.FactoryPairCreate{
				Factory:        hasharry.StringToAddress(factory),
				TokenA:         hasharry.StringToAddress(tokenA),
				TokenB:         hasharry.StringToAddress(tokenB),
				To:             hasharry.StringToAddress(to),
				AmountADesired: amountADesired,
				AmountBDesired: amountBDesired,
				AmountAMin:     amountAMin,
				AmountBMin:     amountBMin,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}
