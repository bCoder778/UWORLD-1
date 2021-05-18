package contractv2

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
)

type ContractType uint
type FunctionType uint

const (
	Exchange_ ContractType = 0
)

const (
	Exchange_Init_ FunctionType = 0
)

type ContractV2 struct {
	Address    hasharry.Address
	CreateHash hasharry.Hash
	Type       ContractType
	Body       IContractV2Body
}

func (c *ContractV2) Bytes() []byte {
	rlpC := &RlpContractV2{
		Address:    c.Address,
		CreateHash: c.CreateHash,
		Type:       c.Type,
		Body:       c.Body.Bytes(),
	}
	bytes, _ := rlp.EncodeToBytes(rlpC)
	return bytes
}

func (c *ContractV2) Verify(function FunctionType) error {
	switch function {
	case Exchange_Init_:
		return fmt.Errorf("exchange %s already exist", c.Address.String())
	}
	return nil
}

type RlpContractV2 struct {
	Address    hasharry.Address
	CreateHash hasharry.Hash
	Type       ContractType
	Body       []byte
}

type IContractV2Body interface {
	Bytes() []byte
}

func DecodeContractV2(bytes []byte) (*ContractV2, error) {
	var rlpContract *RlpContractV2
	if err := rlp.DecodeBytes(bytes, rlpContract); err != nil {
		return nil, err
	}
	var contract = &ContractV2{
		Address:    rlpContract.Address,
		CreateHash: rlpContract.CreateHash,
		Type:       rlpContract.Type,
		Body:       nil,
	}
	switch rlpContract.Type {
	case Exchange_:
		ex, err := DecodeToExchange(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = ex
		return contract, err
	}
	return nil, errors.New("decoding failure")
}
