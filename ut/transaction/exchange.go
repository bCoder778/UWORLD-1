package transaction

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/common/utils"
	"github.com/uworldao/UWORLD/core/runner/exchange_runner"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
	"github.com/uworldao/UWORLD/param"
	"time"
)

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
		TxBody: &types.TxContractV2Body{
			Contract:     hasharry.StringToAddress(contract),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_Init,
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
		TxBody: &types.TxContractV2Body{
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_SetAdmin,
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
		TxBody: &types.TxContractV2Body{
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_SetFeeTo,
			Function: &exchange_func.ExchangeFeeTo{
				Address: hasharry.StringToAddress(feeTo),
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func NewPairCreate(net, from, to, exchange, tokenA, tokenB string, amountADesired, amountBDesired, amountAMin, amountBMin, nonce uint64, note string) (*types.Transaction, error) {
	contract, err := exchange_runner.PairAddress(net, hasharry.StringToAddress(tokenA), hasharry.StringToAddress(tokenB), hasharry.StringToAddress(exchange))
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
		TxBody: &types.TxContractV2Body{
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
		TxBody: &types.TxContractV2Body{
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_ExactIn,
			Function: &exchange_func.ExactIn{
				AmountIn:     amountIn,
				AmountOutMin: amountOutMin,
				Path:         address,
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
		TxBody: &types.TxContractV2Body{
			Contract:     hasharry.StringToAddress(exchange),
			Type:         contractv2.Exchange_,
			FunctionType: contractv2.Exchange_ExactOut,
			Function: &exchange_func.ExactOut{
				AmountOut:   amountOut,
				AmountInMax: amountInMax,
				Path:        address,
				To:          hasharry.StringToAddress(to),
				Deadline:    deadline,
			},
		},
	}
	tx.SetHash()
	return tx, nil
}

func CalculateShortestPath(tokenA, tokenB hasharry.Address, pairs []*types.RpcPair) []string {
	g := utils.NewGraph()
	for _, pair := range pairs {
		g.AddEdge(utils.NewNode(pair.Token0, 0), utils.NewNode(pair.Token1, 0))
	}
	paths, _ := g.FindNodePath(utils.NewNode(tokenA.String(), 0), utils.NewNode(tokenB.String(), 0))
	maxLen := 0
	index := 0
	for i, path := range paths {
		if len(path) < maxLen {
			maxLen = len(path)
			index = i
		}
	}
	tokenPath := []string{}
	for _, node := range paths[index] {
		tokenPath = append(tokenPath, node.String())
	}
	return tokenPath
}
