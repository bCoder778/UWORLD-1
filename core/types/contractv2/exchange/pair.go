package exchange

import (
	"errors"
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/common/math"
)

const MINIMUM_LIQUIDITY = 1e3

type Pair struct {
	Exchange             hasharry.Address
	Token0               hasharry.Address
	Token1               hasharry.Address
	Reserve0             uint64
	Reserve1             uint64
	BlockTimestampLast   uint32
	Price0CumulativeLast uint64
	Price1CumulativeLast uint64
	KLast                uint64
	TotalSupply          uint64
	*liquidityList
}

func NewPair(exchange, token0, token1 hasharry.Address) *Pair {
	return &Pair{
		Exchange:             exchange,
		Token0:               token0,
		Token1:               token1,
		Reserve0:             0,
		Reserve1:             0,
		BlockTimestampLast:   0,
		Price0CumulativeLast: 0,
		Price1CumulativeLast: 0,
		KLast:                0,
		TotalSupply:          0,
		liquidityList:        &liquidityList{},
	}
}

func (p *Pair) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(p)
	return bytes
}

func (p *Pair) Verify() error {
	return nil
}

func (p *Pair) GetReserves() (uint64, uint64, uint32) {
	return p.Reserve0, p.Reserve1, p.BlockTimestampLast
}

func (p *Pair) Mint(address string, number uint64) error {
	value := math.NewMath(p.TotalSupply).Add(number)
	if value.Failed {
		return errors.New("overflow")
	}
	p.TotalSupply = value.Value
	p.liquidityList.mint(address, number)
	return nil
}

func (p *Pair) Update(amount0, amount1 uint64, feeOn bool, blockTime uint64) error {

	blockTimestamp := uint32(blockTime%2 ^ 32)
	timeElapsed := blockTimestamp - p.BlockTimestampLast // overflow is desired
	if timeElapsed > 0 && p.Reserve0 != 0 && p.Reserve1 != 0 {
		// * never overflows, and + overflow is desired
		// 这两个值用于价格预言机
		p.Price0CumulativeLast += p.Reserve1 / p.Reserve0 * uint64(timeElapsed)
		p.Price1CumulativeLast += p.Reserve0 / p.Reserve1 * uint64(timeElapsed)
	}
	p.BlockTimestampLast = blockTimestamp

	value0 := math.NewMath(p.Reserve0).Add(amount0)
	value1 := math.NewMath(p.Reserve1).Add(amount1)
	kLast := value0.Mul(value1.Value)
	if value0.Failed || value1.Failed || kLast.Failed {
		return errors.New("overflow")
	}

	if feeOn {
		p.KLast = kLast.Value
	}
	p.Reserve0 = value0.Value
	p.Reserve1 = value1.Value
	return nil
}

func DecodeToPair(bytes []byte) (*Pair, error) {
	var pair *Pair
	err := rlp.DecodeBytes(bytes, &pair)
	return pair, err
}

type liquidity struct {
	Address string
	Number  uint64
}

type liquidityList []*liquidity

func (l *liquidityList) Get(address string) uint64 {
	for _, liquidity := range *l {
		if liquidity.Address == address {
			return liquidity.Number
		}
	}
	return 0
}

func (l *liquidityList) mint(address string, number uint64) {
	for i, liquidity := range *l {
		if liquidity.Address == address {
			(*l)[i].Number += number
			return
		}
	}
	*l = append(*l, &liquidity{
		Address: address,
		Number:  number,
	})
}

func (l *liquidityList) burn(address string, number uint64) {
	for i, liquidity := range *l {
		if liquidity.Address == address {
			if liquidity.Number >= number {
				(*l)[i].Number -= number
			}
			return
		}
	}
}
