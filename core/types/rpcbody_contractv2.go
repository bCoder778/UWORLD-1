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

type IRCFunction interface {
}

type RpcExchangeInitBody struct {
	FeeToSetter string `json:"feetosetter"`
	FeeTo       string `json:"feeto"`
}
