package types

import (
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
)

type RlpTransaction struct {
	TxHead *TransactionHead
	TxBody []byte
}

type RlpContract struct {
	TxHead *TransactionHead
	TxBody RlpContractBody
}

type RlpContractBody struct {
	Contract     hasharry.Address
	Type         contractv2.ContractType
	FunctionType contractv2.FunctionType
	Function     []byte
	State        ContractState
	Message      string
}

func (rt *RlpTransaction) TranslateToTransaction() *Transaction {
	switch rt.TxHead.TxType {
	case Transfer_:
		var nt *NormalTransactionBody
		rlp.DecodeBytes(rt.TxBody, &nt)
		return &Transaction{
			TxHead: rt.TxHead,
			TxBody: nt,
		}
	case Contract_:
		var ct *ContractBody
		rlp.DecodeBytes(rt.TxBody, &ct)
		return &Transaction{
			TxHead: rt.TxHead,
			TxBody: ct,
		}
	case ContractV2_:
		var ct = &ContractV2Body{}
		var rlpCt *RlpContractBody
		rlp.DecodeBytes(rt.TxBody, &rlpCt)
		switch rlpCt.FunctionType {
		case contractv2.Exchange_Init_:
			var init *exchange_func.ExchangeInitBody
			rlp.DecodeBytes(rlpCt.Function, &init)
			ct.Function = init
		case contractv2.Exchange_SetAdmin_:
			var set *exchange_func.ExchangeAdmin
			rlp.DecodeBytes(rlpCt.Function, &set)
			ct.Function = set
		case contractv2.Exchange_SetFeeTo_:
			var set *exchange_func.ExchangeFeeTo
			rlp.DecodeBytes(rlpCt.Function, &set)
			ct.Function = set
		case contractv2.Pair_Create:
			var create *exchange_func.ExchangePairCreate
			rlp.DecodeBytes(rlpCt.Function, &create)
			ct.Function = create
		}
		rlp.DecodeBytes(rt.TxBody, &ct)
		return &Transaction{
			TxHead: rt.TxHead,
			TxBody: ct,
		}
	case LoginCandidate:
		var nt *LoginTransactionBody
		rlp.DecodeBytes(rt.TxBody, &nt)
		return &Transaction{
			TxHead: rt.TxHead,
			TxBody: nt,
		}
		/*case LogoutCandidate:
			return &Transaction{
				TxHead: rt.TxHead,
				TxBody: &LogoutTransactionBody{},
			}
		case VoteToCandidate:
			var nt *VoteTransactionBody
			rlp.DecodeBytes(rt.TxBody, &nt)
			return &Transaction{
				TxHead: rt.TxHead,
				TxBody: nt,
			}*/
	}
	return nil
}
