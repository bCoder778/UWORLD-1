package contractv2

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2/factory"
)

type ContractType uint
type FunctionType uint

const (
	Factory_ ContractType = 0
	Pair_                 = 1
)

const (
	Factory_Init_     FunctionType = 0
	Factory_SetAdmin_              = 1
	Factory_SetFeeTo_              = 2
	Pair_Create                    = 3
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

func (c *ContractV2) Verify(function FunctionType, sender hasharry.Address) error {
	ex, _ := c.Body.(*factory.Factory)
	switch function {
	case Factory_Init_:
		return fmt.Errorf("factory %s already exist", c.Address.String())
	case Factory_SetAdmin_:
		return ex.VerifySetter(sender)
	case Factory_SetFeeTo_:
		return ex.VerifySetter(sender)
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
	if err := rlp.DecodeBytes(bytes, &rlpContract); err != nil {
		return nil, err
	}
	var contract = &ContractV2{
		Address:    rlpContract.Address,
		CreateHash: rlpContract.CreateHash,
		Type:       rlpContract.Type,
		Body:       nil,
	}
	switch rlpContract.Type {
	case Factory_:
		ex, err := factory.DecodeToFactory(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = ex
		return contract, err
	case Pair_:
		pair, err := factory.DecodeToPair(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = pair
		return contract, err
	}
	return nil, errors.New("decoding failure")
}
