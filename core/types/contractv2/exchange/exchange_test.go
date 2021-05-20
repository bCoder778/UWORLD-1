package exchange

import (
	"fmt"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"testing"
)

type PairTest struct {
	Address string
}

type Test struct {
	List []PairTest
}

func TestNewExchange(t *testing.T) {
	p := PairTest{Address: "123"}
	te := make([]PairTest, 0)
	te = append(te, p)
	test := &Test{List: te}
	bytes, err := rlp.EncodeToBytes(test)
	fmt.Println(err)
	fmt.Println(string(bytes))
}
