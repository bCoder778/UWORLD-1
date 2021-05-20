package contractv2

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types/contractv2/exchange"
)

type ContractType uint
type FunctionType uint

const (
	Exchange_ ContractType = 0
	Pair_                  = 1
)

const (
	Exchange_Init_     FunctionType = 0
	Exchange_SetAdmin_              = 1
	Exchange_SetFeeTo_              = 2
	Exchange_ExactIn                = 3
	Exchange_ExactOut               = 4
	Pair_Create                     = 5
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
	ex, _ := c.Body.(*exchange.Exchange)
	switch function {
	case Exchange_Init_:
		return fmt.Errorf("exchange %s already exist", c.Address.String())
	case Exchange_SetAdmin_:
		return ex.VerifySetter(sender)
	case Exchange_SetFeeTo_:
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
	case Exchange_:
		ex, err := exchange.DecodeToExchange(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = ex
		return contract, err
	case Pair_:
		pair, err := exchange.DecodeToPair(rlpContract.Body)
		if err != nil {
			return nil, err
		}
		contract.Body = pair
		return contract, err
	}
	return nil, errors.New("decoding failure")
}
