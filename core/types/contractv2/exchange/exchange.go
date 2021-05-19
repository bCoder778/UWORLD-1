package exchange

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
	FeeTo    hasharry.Address
	Admin    hasharry.Address
	AllPairs []pair
}

type Exchange struct {
	FeeTo    hasharry.Address
	Admin    hasharry.Address
	Pair     map[hasharry.Address]map[hasharry.Address]hasharry.Address
	AllPairs []pair
}

func NewExchange(admin, feeTo hasharry.Address) *Exchange {
	return &Exchange{
		FeeTo:    admin,
		Admin:    feeTo,
		Pair:     make(map[hasharry.Address]map[hasharry.Address]hasharry.Address),
		AllPairs: make([]pair, 0),
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
		tokenB, token2 := parseKey(pair.key)
		ex.Pair[tokenB] = map[hasharry.Address]hasharry.Address{token2: pair.address}
	}
	return ex, nil
}

func pairKey(tokenB hasharry.Address, token2 hasharry.Address) string {
	str1 := tokenB.String()
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
