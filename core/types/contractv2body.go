package types

import (
	"errors"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

type IFunction interface {
	Verify() error
}

type TxContractV2Body struct {
	Contract     hasharry.Address
	Type         contractv2.ContractType
	FunctionType contractv2.FunctionType
	Function     IFunction
}

func (c *TxContractV2Body) ToAddress() *Receivers {
	recis := NewReceivers()
	return recis
}

func (c *TxContractV2Body) GetAmount() uint64 {
	return 0
}

func (c *TxContractV2Body) GetContract() hasharry.Address {
	return c.Contract
}

func (c *TxContractV2Body) GetName() string {
	return ""
}

func (c *TxContractV2Body) GetAbbr() string {
	return ""
}

func (c *TxContractV2Body) GetIncreaseSwitch() bool {
	return false
}

func (c *TxContractV2Body) GetDescription() string {
	return ""
}

func (c *TxContractV2Body) GetPeerId() []byte {
	return nil
}

func (c *TxContractV2Body) VerifyBody(address hasharry.Address) error {
	if err := c.checkType(); err != nil {
		return err
	}
	if err := c.checkType(); err != nil {
		return err
	}
	return c.Function.Verify()
}

func (c *TxContractV2Body) checkType() error {
	switch c.Type {
	case contractv2.Exchange_:
		switch c.FunctionType {
		case contractv2.Exchange_Init:
			return nil
		case contractv2.Exchange_SetAdmin:
			return nil
		case contractv2.Exchange_SetFeeTo:
			return nil
		case contractv2.Exchange_ExactIn:
			return nil
		case contractv2.Exchange_ExactOut:
			return nil
		}
		return errors.New("invalid contract function type")
	case contractv2.Pair_:
		switch c.FunctionType {
		case contractv2.Pair_AddLiquidity:
			return nil
		case contractv2.Pair_RemoveLiquidity:
			return nil
		}
		return errors.New("invalid contract function type")
	}
	return errors.New("invalid contract type")
}

type ContractState uint8

const (
	Contract_Success ContractState = 0
	Contract_Failed  ContractState = 1
	Contract_Wait    ContractState = 2
)

type ContractV2State struct {
	State ContractState
	Event []*Event
	Error string
}

func (c *ContractV2State) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(c)
	return bytes
}

func DecodeContractV2State(bytes []byte) (*ContractV2State, error) {
	var c *ContractV2State
	err := rlp.DecodeBytes(bytes, &c)
	return c, err
}

type EventType uint32

const (
	Event_Transfer EventType = 0
	Event_Mint     EventType = 1
	Event_Burn     EventType = 2
)

type Event struct {
	EventType EventType
	From      hasharry.Address
	To        hasharry.Address
	Token     hasharry.Address
	Amount    uint64
	Height    uint64
}
