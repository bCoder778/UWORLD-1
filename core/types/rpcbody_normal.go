package types

type RpcTransferBody struct {
	Contract string `json:"contract"`
	To       string `json:"to"`
	Amount   uint64 `json:"amount"`
}

func (rnb *RpcTransferBody) ToBytes() []byte {
	return []byte(rnb.To)
}

type RpcTransferV2Body struct {
	Contract  string        `json:"contract"`
	Receivers []RpcReceiver `json:"receivers"`
}

type RpcReceiver struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}
