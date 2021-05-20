package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2"
	"github.com/uworldao/UWORLD/core/types/functionbody/exchange_func"
)

type IRpcTransactionBody interface {
}

type RpcTransactionHead struct {
	TxHash     string          `json:"txhash"`
	TxType     TransactionType `json:"txtype"`
	From       string          `json:"from"`
	Nonce      uint64          `json:"nonce"`
	Fees       uint64          `json:"fees"`
	Time       uint64          `json:"time"`
	Note       string          `json:"note"`
	SignScript *RpcSignScript  `json:"signscript"`
}

type RpcTransaction struct {
	TxHead *RpcTransactionHead `json:"txhead"`
	TxBody IRpcTransactionBody `json:"txbody"`
}

type RpcTransactionConfirmed struct {
	TxHead    *RpcTransactionHead `json:"txhead"`
	TxBody    IRpcTransactionBody `json:"txbody"`
	Height    uint64              `json:"height"`
	Confirmed bool                `json:"confirmed"`
}

type RpcSignScript struct {
	Signature string `json:"signature"`
	PubKey    string `json:"pubkey"`
}

func (th *RpcTransactionHead) FromBytes() []byte {
	return []byte(th.From)
}

func TranslateRpcTxToTx(rpcTx *RpcTransaction) (*Transaction, error) {
	var err error
	txHash, err := hasharry.StringToHash(rpcTx.TxHead.TxHash)
	if err != nil {
		return nil, err
	}
	signScript, err := TranslateRpcSignScriptToSignScript(rpcTx.TxHead.SignScript)
	if err != nil {
		return nil, err
	}
	var txBody ITransactionBody
	switch rpcTx.TxHead.TxType {
	case Transfer_:
		body := &RpcNormalTransactionBody{}
		bytes, err := json.Marshal(rpcTx.TxBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		txBody, err = translateRpcNormalBodyToBody(body)
	case Contract_:
		body := &RpcContractTransactionBody{}
		bytes, err := json.Marshal(rpcTx.TxBody)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
		txBody, err = translateRpcContractBodyToBody(body)
	case ContractV2_:
		txBody, err = translateRpcContractV2BodyToBody(rpcTx.TxBody)
		/*case types.LoginCandidate:
			txBody, err = translateRpcLoginBodyToBody(rpcTx.LoginBody)
		case types.LogoutCandidate:
			txBody = &types.LogoutTransactionBody{}
		case types.VoteToCandidate:
			txBody, err = translateRpcVoteBodyToBody(rpcTx.VoteBody)*/
	}
	tx := &Transaction{
		TxHead: &TransactionHead{
			TxHash:     txHash,
			TxType:     rpcTx.TxHead.TxType,
			From:       hasharry.StringToAddress(rpcTx.TxHead.From),
			Nonce:      rpcTx.TxHead.Nonce,
			Fees:       rpcTx.TxHead.Fees,
			Time:       rpcTx.TxHead.Time,
			Note:       rpcTx.TxHead.Note,
			SignScript: signScript,
		},
		TxBody: txBody,
	}
	return tx, nil
}

func TranslateTxToRpcTx(tx *Transaction) (*RpcTransaction, error) {
	var err error
	rpcTx := &RpcTransaction{
		TxHead: &RpcTransactionHead{
			TxHash: tx.Hash().String(),
			TxType: tx.GetTxType(),
			From:   addressToString(tx.From()),
			Nonce:  tx.GetNonce(),
			Fees:   tx.GetFees(),
			Time:   tx.GetTime(),
			Note:   tx.GetNote(),
			SignScript: &RpcSignScript{
				Signature: hex.EncodeToString(tx.GetSignScript().Signature),
				PubKey:    hex.EncodeToString(tx.GetSignScript().PubKey),
			}},
		TxBody: nil,
	}
	switch tx.GetTxType() {
	case Transfer_:
		rpcTx.TxBody = &RpcNormalTransactionBody{
			Contract: tx.GetTxBody().GetContract().String(),
			To:       tx.GetTxBody().ToAddress().String(),
			Amount:   tx.GetTxBody().GetAmount(),
		}
	case Contract_:
		rpcTx.TxBody = &RpcContractTransactionBody{
			Contract:    tx.GetTxBody().GetContract().String(),
			To:          tx.GetTxBody().ToAddress().String(),
			Name:        tx.GetTxBody().GetName(),
			Abbr:        tx.GetTxBody().GetAbbr(),
			Description: tx.GetTxBody().GetDescription(),
			Increase:    tx.GetTxBody().GetIncreaseSwitch(),
			Amount:      tx.GetTxBody().GetAmount(),
		}
	case ContractV2_:
		body, ok := tx.GetTxBody().(*ContractV2Body)
		if !ok {
			return nil, errors.New("wrong transaction body")
		}
		rpcTx.TxBody, err = translateContractV2ToRpcContractV2(body)
		if err != nil {
			return nil, err
		}
	case LoginCandidate:
		rpcTx.TxBody = &RpcLoginTransactionBody{
			PeerId: string(tx.GetTxBody().GetPeerId()),
		}
		/*case types.LogoutCandidate:
			rpcTx.LogoutBody = &RpcLogoutTransactionBody{}
		case types.VoteToCandidate:
			rpcTx.VoteBody = &RpcVoteTransactionBody{To: tx.GetTxBody().ToAddress().String()}
		*/
	}

	return rpcTx, nil
}

func TranslateContractV2TxToRpcTx(tx *Transaction, state *ContractV2State) (*RpcTransaction, error) {
	var err error
	rpcTx := &RpcTransaction{
		TxHead: &RpcTransactionHead{
			TxHash: tx.Hash().String(),
			TxType: tx.GetTxType(),
			From:   addressToString(tx.From()),
			Nonce:  tx.GetNonce(),
			Fees:   tx.GetFees(),
			Time:   tx.GetTime(),
			Note:   tx.GetNote(),
			SignScript: &RpcSignScript{
				Signature: hex.EncodeToString(tx.GetSignScript().Signature),
				PubKey:    hex.EncodeToString(tx.GetSignScript().PubKey),
			}},
		TxBody: nil,
	}
	switch tx.GetTxType() {
	case ContractV2_:
		body, ok := tx.GetTxBody().(*ContractV2Body)
		if !ok {
			return nil, errors.New("wrong transaction body")
		}
		rpcTx.TxBody, err = translateToRpcContractV2WithState(body, state)
		if err != nil {
			return nil, err
		}
	}

	return rpcTx, nil
}

func rpcFunction(body *ContractV2Body) (IRCFunction, error) {
	var function IRCFunction
	switch body.FunctionType {
	case contractv2.Exchange_Init_:
		funcBody, ok := body.Function.(*exchange_func.ExchangeInitBody)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeInitBody{
			Admin: funcBody.Admin.String(),
			FeeTo: funcBody.FeeTo.String(),
		}
	case contractv2.Exchange_SetAdmin_:
		funcBody, ok := body.Function.(*exchange_func.ExchangeAdmin)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeSetAdminBody{
			Address: funcBody.Address.String(),
		}
	case contractv2.Exchange_SetFeeTo_:
		funcBody, ok := body.Function.(*exchange_func.ExchangeFeeTo)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangeSetFeeToBody{
			Address: funcBody.Address.String(),
		}
	case contractv2.Pair_Create:
		funcBody, ok := body.Function.(*exchange_func.ExchangePairCreate)
		if !ok {
			return nil, errors.New("wrong function body")
		}
		function = &RpcExchangePairCreate{
			Exchange:       funcBody.Exchange.String(),
			TokenA:         funcBody.TokenA.String(),
			TokenB:         funcBody.TokenB.String(),
			To:             funcBody.To.String(),
			AmountADesired: Amount(funcBody.AmountADesired).ToCoin(),
			AmountBDesired: Amount(funcBody.AmountBDesired).ToCoin(),
			AmountAMin:     Amount(funcBody.AmountAMin).ToCoin(),
			AmountBMin:     Amount(funcBody.AmountBMin).ToCoin(),
		}
	}
	return function, nil
}

func translateToRpcContractV2WithState(body *ContractV2Body, contractState *ContractV2State) (*RpcContractV2BodyWithState, error) {
	var state = Contract_Wait
	var message string
	if contractState != nil {
		state = contractState.State
		message = contractState.Message
	}

	funcBody, err := rpcFunction(body)
	if err != nil {
		return nil, err
	}
	return &RpcContractV2BodyWithState{
		Contract:     body.Contract.String(),
		Type:         body.Type,
		FunctionType: body.FunctionType,
		Function:     funcBody,
		State:        state,
		Message:      message,
	}, nil
}

func translateContractV2ToRpcContractV2(body *ContractV2Body) (*RpcContractV2TransactionBody, error) {
	funcBody, err := rpcFunction(body)
	if err != nil {
		return nil, err
	}
	return &RpcContractV2TransactionBody{
		Contract:     body.Contract.String(),
		Type:         body.Type,
		FunctionType: body.FunctionType,
		Function:     funcBody,
	}, nil
}

func TranslateRpcSignScriptToSignScript(rpcSignScript *RpcSignScript) (*SignScript, error) {
	if rpcSignScript == nil {
		return nil, ErrNoSignature
	}
	if rpcSignScript.Signature == "" || rpcSignScript.PubKey == "" {
		return nil, ErrWrongSignature
	}
	signature, err := hex.DecodeString(rpcSignScript.Signature)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(rpcSignScript.PubKey)
	if err != nil {
		return nil, err
	}
	return &SignScript{
		Signature: signature,
		PubKey:    pubKey,
	}, nil
}

func translateRpcNormalBodyToBody(rpcBody *RpcNormalTransactionBody) (*NormalTransactionBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong transaction body")
	}

	return &NormalTransactionBody{
		Contract: hasharry.StringToAddress(rpcBody.Contract),
		To:       hasharry.StringToAddress(rpcBody.To),
		Amount:   rpcBody.Amount,
	}, nil
}

func translateRpcContractBodyToBody(rpcBody *RpcContractTransactionBody) (*ContractBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong contract transaction body")
	}

	return &ContractBody{
		Contract:       hasharry.StringToAddress(rpcBody.Contract),
		To:             hasharry.StringToAddress(rpcBody.To),
		Abbr:           rpcBody.Abbr,
		IncreaseSwitch: rpcBody.Increase,
		Name:           rpcBody.Name,
		Description:    rpcBody.Description,
		Amount:         rpcBody.Amount,
	}, nil
}

func translateRpcContractV2BodyToBody(rpcBody IRpcTransactionBody) (*ContractV2Body, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong contract transaction body")
	}
	body := &RpcContractV2TransactionBody{}
	bytes, err := json.Marshal(rpcBody)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, body)
	if err != nil {
		return nil, err
	}
	switch body.FunctionType {
	case contractv2.Exchange_Init_:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		init := &RpcExchangeInitBody{
			Admin: "",
			FeeTo: "",
		}
		err = json.Unmarshal(bytes, init)
		if err != nil {
			return nil, err
		}
		return &ContractV2Body{
			Contract:     hasharry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeInitBody{
				Admin: hasharry.StringToAddress(init.Admin),
				FeeTo: hasharry.StringToAddress(init.FeeTo),
			},
		}, nil
	case contractv2.Exchange_SetAdmin_:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		setBody := &RpcExchangeSetAdminBody{
			Address: "",
		}
		err = json.Unmarshal(bytes, setBody)
		if err != nil {
			return nil, err
		}
		return &ContractV2Body{
			Contract:     hasharry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeAdmin{
				Address: hasharry.StringToAddress(setBody.Address),
			},
		}, nil
	case contractv2.Exchange_SetFeeTo_:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		setBody := &RpcExchangeSetFeeToBody{
			Address: "",
		}
		err = json.Unmarshal(bytes, setBody)
		if err != nil {
			return nil, err
		}
		return &ContractV2Body{
			Contract:     hasharry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangeFeeTo{
				Address: hasharry.StringToAddress(setBody.Address),
			},
		}, nil
	case contractv2.Pair_Create:
		bytes, err := json.Marshal(body.Function)
		if err != nil {
			return nil, err
		}
		createBody := &RpcExchangePairCreate{}
		err = json.Unmarshal(bytes, createBody)
		if err != nil {
			return nil, err
		}
		amountADesired, _ := NewAmount(createBody.AmountADesired)
		amountBDesired, _ := NewAmount(createBody.AmountBDesired)
		amountAMin, _ := NewAmount(createBody.AmountAMin)
		amountBMin, _ := NewAmount(createBody.AmountBMin)
		return &ContractV2Body{
			Contract:     hasharry.StringToAddress(body.Contract),
			Type:         body.Type,
			FunctionType: body.FunctionType,
			Function: &exchange_func.ExchangePairCreate{
				Exchange:       hasharry.StringToAddress(createBody.Exchange),
				TokenA:         hasharry.StringToAddress(createBody.TokenA),
				TokenB:         hasharry.StringToAddress(createBody.TokenB),
				To:             hasharry.StringToAddress(createBody.To),
				AmountADesired: amountADesired,
				AmountBDesired: amountBDesired,
				AmountAMin:     amountAMin,
				AmountBMin:     amountBMin,
			},
		}, nil
	}
	return nil, errors.New("wrong transaction body")
}

func translateRpcLoginBodyToBody(rpcBody *RpcLoginTransactionBody) (*LoginTransactionBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong transaction body")
	}
	loginTx := &LoginTransactionBody{}
	copy(loginTx.PeerId[:], rpcBody.PeerIdBytes())
	return loginTx, nil
}

func translateRpcVoteBodyToBody(rpcBody *RpcVoteTransactionBody) (*VoteTransactionBody, error) {
	if rpcBody == nil {
		return nil, errors.New("wrong transaction body")
	}

	return &VoteTransactionBody{To: hasharry.StringToAddress(rpcBody.To)}, nil
}

func addressToString(address hasharry.Address) string {
	if address.IsEqual(hasharry.StringToAddress(CoinBase)) {
		return CoinBase
	}
	return address.String()
}
