package exchange_func

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type ExchangeInitBody struct {
	Admin hasharry.Address
	FeeTo hasharry.Address
}

func (e *ExchangeInitBody) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Admin.String()); !ok {
		return errors.New("wrong admin address")
	}
	feeTo := e.FeeTo.String()
	if feeTo != "" {
		if ok := ut.CheckUWDAddress(param.Net, feeTo); !ok {
			return errors.New("wrong feeTo address")
		}
	}
	return nil
}

type ExchangeAdmin struct {
	Address hasharry.Address
}

func (e *ExchangeAdmin) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Address.String()); !ok {
		return errors.New("wrong admin address")
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
