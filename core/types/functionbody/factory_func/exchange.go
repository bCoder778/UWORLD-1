package factory_func

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type FactoryInitBody struct {
	Admin hasharry.Address
	FeeTo hasharry.Address
}

func (e *FactoryInitBody) Verify() error {
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

type FactoryAdmin struct {
	Address hasharry.Address
}

func (e *FactoryAdmin) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Address.String()); !ok {
		return errors.New("wrong admin address")
	}
	return nil
}

type FactoryFeeTo struct {
	Address hasharry.Address
}

func (e *FactoryFeeTo) Verify() error {
	if ok := ut.CheckUWDAddress(param.Net, e.Address.String()); !ok {
		return errors.New("wrong feeTo address")
	}
	return nil
}
