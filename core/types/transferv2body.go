package types

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type TransferV2Body struct {
	Contract  hasharry.Address
	Receivers *Receivers
}

type Receiver struct {
	Address hasharry.Address
	Amount  uint64
}

func (tb *TransferV2Body) ToAddress() *Receivers {
	return tb.Receivers
}

func (tb *TransferV2Body) GetAmount() uint64 {
	var sum uint64
	for _, re := range tb.Receivers.List {
		sum += re.Amount
	}
	return sum
}

func (tb *TransferV2Body) GetContract() hasharry.Address {
	return tb.Contract
}

func (tb *TransferV2Body) GetName() string {
	return ""
}

func (tb *TransferV2Body) GetAbbr() string {
	return ""
}

func (tb *TransferV2Body) GetIncreaseSwitch() bool {
	return false
}

func (tb *TransferV2Body) GetDescription() string {
	return ""
}

func (tb *TransferV2Body) GetPeerId() []byte {
	return nil
}

func (tb *TransferV2Body) VerifyBody(from hasharry.Address) error {
	if len(tb.Receivers.List) > param.MaximumReceiver {
		return fmt.Errorf("the maximum number of receive addresses is %d", param.MaximumReceiver)
	}
	if len(tb.Receivers.List) == 0 {
		return fmt.Errorf("no receivers")
	}
	if err := tb.Receivers.CheckAddress(); err != nil {
		return err
	}
	if !tb.Contract.IsEqual(param.Token) {
		if !ut.IsValidContractAddress(param.Net, tb.Contract.String()) {
			return errors.New("token address verification failed")
		}
	}
	return nil
}

type Receivers struct {
	List []*Receiver
}

func NewReceivers() *Receivers {
	return &Receivers{
		List: make([]*Receiver, 0),
	}
}

func (r *Receivers) Add(address hasharry.Address, amount uint64) {
	r.List = append(r.List, &Receiver{
		Address: address,
		Amount:  amount,
	})
}

func (r *Receivers) CheckAddress() error {
	for _, re := range r.List {
		if !ut.CheckUWDAddress(param.Net, re.Address.String()) {
			return fmt.Errorf("receive address %s verification failed", re.Address.String())
		}
	}
	return nil
}

func (r *Receivers) ReceiverList() []*Receiver {
	return r.List
}
