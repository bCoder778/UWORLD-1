package types

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2"
)

type IFunction interface {
	Verify() error
}

type ContractV2Body struct {
	Contract     hasharry.Address
	Type         contractv2.ContractType
	FunctionType contractv2.FunctionType
	Function     IFunction
	State        ContractState
	Message      string
}

func (c *ContractV2Body) ToAddress() hasharry.Address {
	return hasharry.Address{}
}

func (c *ContractV2Body) GetAmount() uint64 {
	return 0
}

func (c *ContractV2Body) GetContract() hasharry.Address {
	return c.Contract
}

func (c *ContractV2Body) GetName() string {
	return ""
}

func (c *ContractV2Body) GetAbbr() string {
	return ""
}

func (c *ContractV2Body) GetIncreaseSwitch() bool {
	return false
}

func (c *ContractV2Body) GetDescription() string {
	return ""
}

func (c *ContractV2Body) GetPeerId() []byte {
	return nil
}

func (c *ContractV2Body) VerifyBody(address hasharry.Address) error {
	if err := c.checkType(); err != nil {
		return err
	}
	if err := c.checkType(); err != nil {
		return err
	}
	return c.Function.Verify()
}

func (c *ContractV2Body) checkType() error {
	switch c.Type {
	case contractv2.Exchange_:
		switch c.FunctionType {
		case contractv2.Exchange_Init_:
			return nil
		case contractv2.Exchange_SetAdmin_:
			return nil
		case contractv2.Exchange_SetFeeTo_:
			return nil
		}
		return errors.New("invalid contract function type")
	case contractv2.Pair_:
		switch c.FunctionType {
		case contractv2.Pair_Create:
			return nil
		}
		return errors.New("invalid contract function type")
	}
	return errors.New("invalid contract type")
}

type ContractState uint8

const (
	Contract_Wait    ContractState = 0
	Contract_Success               = 1
	Contract_Failed                = 2
)
