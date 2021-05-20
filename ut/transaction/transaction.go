package transaction

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/runner/exchange_runner"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
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

func NewExchange(net, from, admin, feeTo string, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := exchange_runner.ExchangeAddress(net, from, nonce)
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
			Function: &exchange_func.ExchangeInitBody{
				Admin: hasharry.StringToAddress(admin),
				FeeTo: hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSetAdmin(from, exchange, admin string, nonce uint64, note string) (*types.Transaction, error) {
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
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_SetAdmin_,
			Function: &exchange_func.ExchangeAdmin{
				Address: hasharry.StringToAddress(admin),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSetFeeTo(from, exchange, feeTo string, nonce uint64, note string) (*types.Transaction, error) {
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
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_SetFeeTo_,
			Function: &exchange_func.ExchangeFeeTo{
				Address: hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewPairCreate(net, from, to, exchange, tokenA, tokenB string, amountADesired, amountBDesired, amountAMin, amountBMin, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := exchange_runner.PairAddress(net, tokenA, tokenB)
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
			Function: &exchange_func.ExchangePairCreate{
				Exchange:       hasharry.StringToAddress(exchange),
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

func NewSwapExactIn(from, to, exchange string, amountIn, amountOutMin uint64, path []string, deadline, nonce uint64, note string) (*types.Transaction, error) {
	address := make([]hasharry.Address, 0)
	for _, addr := range path {
		address = append(address, hasharry.StringToAddress(addr))
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
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_ExactIn,
			Function: &exchange_func.ExactIn{
				AmountIn:     amountIn,
				AmountOutMin: amountOutMin,
				Address:      address,
				To:           hasharry.StringToAddress(to),
				Deadline:     deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewSwapExactOut(from, to, exchange string, amountOut, amountInMax uint64, path []string, deadline, nonce uint64, note string) (*types.Transaction, error) {
	address := make([]hasharry.Address, 0)
	for _, addr := range path {
		address = append(address, hasharry.StringToAddress(addr))
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
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_ExactOut,
			Function: &exchange_func.ExactOut{
				AmountOut:   amountOut,
				AmountInMax: amountInMax,
				Address:     address,
				To:          hasharry.StringToAddress(to),
				Deadline:    deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}
