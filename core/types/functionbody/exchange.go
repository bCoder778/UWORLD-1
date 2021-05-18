package functionbody

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type ExchangeInitBody struct {
	FeeToSetter hasharry.Address
	FeeTo       hasharry.Address
}

func (e *ExchangeInitBody) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.FeeToSetter.String()); !ok {
		return errors.New("wrong feeToSetter address")
	}
	feeTo := e.FeeTo.String()
	if feeTo != "" {
		if ok := ut.CheckUWDAddress(param.Net, feeTo); !ok {
			return errors.New("wrong feeTo address")
		}
	}
	return nil
}

type ExchangeFeeToSetter struct {
	Address hasharry.Address
}

func (e *ExchangeFeeToSetter) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Address.String()); !ok {
		return errors.New("wrong feeToSetter address")
	}
	return nil
}

type ExchangeFeeTo struct {
	Address hasharry.Address
}

func (e *ExchangeFeeTo) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Address.String()); !ok {
		return errors.New("wrong feeTo address")
	}
	return nil
}
