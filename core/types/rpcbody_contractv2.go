package types

import (
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

type RpcContractV2TransactionBody struct {
	Contract     string                  `json:"contract"`
	Type         contractv2.ContractType `json:"type"`
	FunctionType contractv2.FunctionType `json:"functiontype"`
	Function     IRCFunction             `json:"function"`
}

type RpcContractV2BodyWithState struct {
	Contract     string                  `json:"contract"`
	Type         contractv2.ContractType `json:"type"`
	FunctionType contractv2.FunctionType `json:"functiontype"`
	Function     IRCFunction             `json:"function"`
	State        ContractState           `json:"state"`
	Message      string                  `json:"message"`
}

type IRCFunction interface {
}

type RpcExchangeInitBody struct {
	Admin string `json:"admin"`
	FeeTo string `json:"feeto"`
}

type RpcExchangeSetAdminBody struct {
	Address string `json:"address"`
}

type RpcExchangeSetFeeToBody struct {
	Address string `json:"address"`
}

type RpcExchangePairCreate struct {
	Exchange       string  `json:"exchange"`
	TokenA         string  `json:"tokenA"`
	TokenB         string  `json:"tokenB"`
	To             string  `json:"to"`
	AmountADesired float64 `json:"amountADesired"`
	AmountBDesired float64 `json:"amountBDesired"`
	AmountAMin     float64 `json:"amountAMin"`
	AmountBMin     float64 `json:"amountBMin"`
}
