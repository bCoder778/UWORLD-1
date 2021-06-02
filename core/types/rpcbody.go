package types

import "github.com/uworldao/UWORLD/common/hasharry"

type RpcBody struct {
	Transactions []*RpcTransaction `json:"transactions"`
}

type getContractV2State func(hasharry.Hash) (*ContractV2State, error)

func TranslateBodyToRpcBody(body *Body, stateFunc getContractV2State) (*RpcBody, error) {
	var rpcTxs []*RpcTransaction
	var rpcTx *RpcTransaction
	var err error
	for _, tx := range body.Transactions {
		if tx.GetTxType() == ContractV2_ {
			state, _ := stateFunc(tx.Hash())
			rpcTx, err = TranslateContractV2TxToRpcTx(tx.(*Transaction), state)
			if err != nil {
				return nil, err
			}
		} else {
			rpcTx, err = TranslateTxToRpcTx(tx.(*Transaction))
			if err != nil {
				return nil, err
			}
		}
		rpcTxs = append(rpcTxs, rpcTx)
	}
	return &RpcBody{rpcTxs}, nil
}
