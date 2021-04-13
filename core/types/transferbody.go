package types

import (
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

const maxContractLength = 50

// Ordinary transfer transaction body
type TransferBody struct {
	Contract hasharry.Address
	To       hasharry.Address
	Amount   uint64
}

func (nt *TransferBody) ToAddress() *Receivers {
	recis := NewReceivers()
	recis.Add(nt.To, nt.Amount)
	return recis
}

func (nt *TransferBody) GetAmount() uint64 {
	return nt.Amount
}

func (nt *TransferBody) GetContract() hasharry.Address {
	return nt.Contract
}

func (nt *TransferBody) GetName() string {
	return ""
}

func (nt *TransferBody) GetAbbr() string {
	return ""
}

func (nt *TransferBody) GetIncreaseSwitch() bool {
	return false
}

func (nt *TransferBody) GetDescription() string {
	return ""
}

func (nt *TransferBody) GetPeerId() []byte {
	return nil
}

func (nt *TransferBody) VerifyBody(from hasharry.Address) error {
	if err := nt.verifyContract(); err != nil {
		return err
	}
	if err := nt.verifyTo(); err != nil {
		return err
	}
	return nil
}

func (nt *TransferBody) verifyTo() error {
	if !ut.CheckUWDAddress(param.Net, nt.To.String()) {
		return ErrAddress
	}
	return nil
}

func (nt *TransferBody) verifyContract() error {
	if !ut.IsValidContractAddress(param.Net, nt.Contract.String()) {
		return ErrContractAddr
	}
	return nil
}
