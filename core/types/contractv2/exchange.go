package contractv2

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"strings"
)

type pair struct {
	key     string
	address hasharry.Address
}

type RlpExchange struct {
	FeeTo       hasharry.Address
	FeeToSetter hasharry.Address
	AllPairs    []pair
}

type Exchange struct {
	FeeTo       hasharry.Address
	FeeToSetter hasharry.Address
	Pair        map[hasharry.Address]map[hasharry.Address]hasharry.Address
	AllPairs    []pair
}

func NewExchange(feeToSetter, feeTo hasharry.Address) *Exchange {
	return &Exchange{
		FeeTo:       feeToSetter,
		FeeToSetter: feeTo,
		Pair:        make(map[hasharry.Address]map[hasharry.Address]hasharry.Address),
		AllPairs:    make([]pair, 0),
	}
}

func (e *Exchange) SetFeeTo(address hasharry.Address, sender hasharry.Address) error {
	if !e.FeeToSetter.IsEqual(sender) {

		return errors.New("forbidden")
	}
	e.FeeTo = address
	return nil
}

func (e *Exchange) SetFeeToSetter(address hasharry.Address, sender hasharry.Address) error {
	if !e.FeeToSetter.IsEqual(sender) {
		return errors.New("forbidden")
	}
	e.FeeToSetter = address
	return nil
}

func (e *Exchange) Bytes() []byte {
	elpEx := &RlpExchange{
		FeeTo:       e.FeeTo,
		FeeToSetter: e.FeeToSetter,
		AllPairs:    e.AllPairs,
	}
	bytes, _ := rlp.EncodeToBytes(elpEx)
	return bytes
}

func DecodeToExchange(bytes []byte) (*Exchange, error) {
	var rlpEx *RlpExchange
	if err := rlp.DecodeBytes(bytes, &rlpEx); err != nil {
		return nil, err
	}
	ex := NewExchange(rlpEx.FeeToSetter, rlpEx.FeeTo)
	ex.AllPairs = rlpEx.AllPairs
	for _, pair := range rlpEx.AllPairs {
		token1, token2 := parseKey(pair.key)
		ex.Pair[token1] = map[hasharry.Address]hasharry.Address{token2: pair.address}
	}
	return ex, nil
}

func pairKey(token1 hasharry.Address, token2 hasharry.Address) string {
	str1 := token1.String()
	str2 := token2.String()
	if strings.Compare(str1, str2) > 0 {
		return fmt.Sprintf("%s-%s", str1, str2)
	} else {
		return fmt.Sprintf("%s-%s", str2, str1)
	}
}

func parseKey(key string) (hasharry.Address, hasharry.Address) {
	strList := strings.Split(key, "-")
	if len(strList) != 2 {
		return hasharry.Address{}, hasharry.Address{}
	}
	return hasharry.StringToAddress(strList[0]), hasharry.StringToAddress(strList[1])
}
