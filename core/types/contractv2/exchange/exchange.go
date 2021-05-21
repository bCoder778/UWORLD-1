package exchange

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"strings"
)

type PairAddress struct {
	Key     string
	Address hasharry.Address
}

type RlpExchange struct {
	FeeTo    hasharry.Address
	Admin    hasharry.Address
	AllPairs []PairAddress
}

type Exchange struct {
	FeeTo    hasharry.Address
	Admin    hasharry.Address
	Pair     map[hasharry.Address]map[hasharry.Address]hasharry.Address
	AllPairs []PairAddress
}

func NewExchange(admin, feeTo hasharry.Address) *Exchange {
	return &Exchange{
		FeeTo:    admin,
		Admin:    feeTo,
		Pair:     make(map[hasharry.Address]map[hasharry.Address]hasharry.Address),
		AllPairs: make([]PairAddress, 0),
	}
}

func (e *Exchange) SetFeeTo(address hasharry.Address, sender hasharry.Address) error {
	if err := e.VerifySetter(sender); err != nil {
		return err
	}
	e.FeeTo = address
	return nil
}

func (e *Exchange) SetAdmin(address hasharry.Address, sender hasharry.Address) error {
	if err := e.VerifySetter(sender); err != nil {
		return err
	}
	e.Admin = address
	return nil
}

func (e *Exchange) VerifySetter(sender hasharry.Address) error {
	if !e.Admin.IsEqual(sender) {
		return errors.New("forbidden")
	}
	return nil
}

func (e *Exchange) Exist(token0, token1 hasharry.Address) bool {
	token1Map, ok := e.Pair[token0]
	if ok {
		_, ok := token1Map[token1]
		return ok
	}
	return false
}

func (e *Exchange) PairAddress(token0, token1 hasharry.Address) hasharry.Address {
	token1Map, ok := e.Pair[token0]
	if ok {
		address, _ := token1Map[token1]
		return address
	}
	return hasharry.Address{}
}

func (e *Exchange) AddPair(token0, token1, address hasharry.Address) {
	e.Pair[token0] = map[hasharry.Address]hasharry.Address{token1: address}
	e.AllPairs = append(e.AllPairs, PairAddress{
		Key:     pairKey(token0, token1),
		Address: address,
	})
}

func (e *Exchange) Bytes() []byte {
	elpEx := &RlpExchange{
		FeeTo:    e.FeeTo,
		Admin:    e.Admin,
		AllPairs: e.AllPairs,
	}
	bytes, _ := rlp.EncodeToBytes(elpEx)
	return bytes
}

func DecodeToExchange(bytes []byte) (*Exchange, error) {
	var rlpEx *RlpExchange
	if err := rlp.DecodeBytes(bytes, &rlpEx); err != nil {
		return nil, err
	}
	ex := NewExchange(rlpEx.Admin, rlpEx.FeeTo)
	ex.AllPairs = rlpEx.AllPairs
	for _, pair := range rlpEx.AllPairs {
		tokenB, token2 := parseKey(pair.Key)
		ex.Pair[tokenB] = map[hasharry.Address]hasharry.Address{token2: pair.Address}
	}
	return ex, nil
}

func pairKey(token0 hasharry.Address, token1 hasharry.Address) string {
	return fmt.Sprintf("%s-%s", token0.String(), token1.String())
}

func parseKey(key string) (hasharry.Address, hasharry.Address) {
	strList := strings.Split(key, "-")
	if len(strList) != 2 {
		return hasharry.Address{}, hasharry.Address{}
	}
	return hasharry.StringToAddress(strList[0]), hasharry.StringToAddress(strList[1])
}
