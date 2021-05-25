package exchange_func

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
	"github.com/uworldao/UWORLD/ut"
)

type ExactIn struct {
	AmountIn     uint64
	AmountOutMin uint64
	Path         []hasharry.Address
	To           hasharry.Address
	Deadline     uint64
}

func (e *ExactIn) Verify() error {
	if !ut.CheckUWDAddress(param.Net, e.To.String()) {
		return errors.New("wrong to address")
	}
	for _, addr := range e.Path {
		if !ut.IsValidContractAddress(param.Net, addr.String()) {
			return fmt.Errorf("wrong path address %s", addr.String())
		}
	}
	return nil
}

type ExactOut struct {
	AmountOut   uint64
	AmountInMax uint64
	Path        []hasharry.Address
	To          hasharry.Address
	Deadline    uint64
}

func (e *ExactOut) Verify() error {
	if !ut.CheckUWDAddress(param.Net, e.To.String()) {
		return errors.New("wrong to address")
	}
	for _, addr := range e.Path {
		if !ut.IsValidContractAddress(param.Net, addr.String()) {
			return fmt.Errorf("wrong path address %s", addr.String())
		}
	}
	return nil
}
