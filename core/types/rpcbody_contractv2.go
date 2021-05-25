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

type RpcExchangeExactInBody struct {
	AmountIn     uint64   `json:"amountin"`
	AmountOutMin uint64   `json:"amountoutmin"`
	Path         []string `json:"path"`
	To           string   `json:"to"`
	Deadline     uint64   `json:"deadline"`
}

type RpcExchangeExactOutBody struct {
	AmountOut   uint64   `json:"amountout"`
	AmountInMax uint64   `json:"amountinmax"`
	Path        []string `json:"path"`
	To          string   `json:"to"`
	Deadline    uint64   `json:"deadline"`
}

type RpcExchangePairCreate struct {
	Exchange       string  `json:"exchange"`
	TokenA         string  `json:"tokenA"`
	TokenB         string  `json:"tokenB"`
	To             string  `json:"to"`
	AmountADesired float64 `json:"amountadesired"`
	AmountBDesired float64 `json:"amountbdesired"`
	AmountAMin     float64 `json:"amountamin"`
	AmountBMin     float64 `json:"amountbmin"`
}

type RpcPair struct {
	Address  string `json:"address"`
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
	Reserve0 uint64 `json:"reserve0"`
	Reserve1 uint64 `json:"reserve1"`
}
