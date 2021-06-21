package exchange

import (
	"github.com/uworldao/UWORLD/common/encode/rlp"
	"github.com/uworldao/UWORLD/common/hasharry"
	"math/big"
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
	KLast                *big.Int
	TotalSupply          uint64
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
		KLast:                big.NewInt(0),
		TotalSupply:          0,
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

func (p *Pair) Mint(address string, number uint64) {
	p.TotalSupply = p.TotalSupply + number
}

func (p *Pair) Update(balance0, balance1, _reserve0, _reserve1, blockTime uint64) {
	blockTimestamp := uint32(blockTime%2 ^ 32)
	timeElapsed := blockTimestamp - p.BlockTimestampLast // overflow is desired
	if timeElapsed > 0 && _reserve0 != 0 && _reserve1 != 0 {
		// * never overflows, and + overflow is desired
		// 这两个值用于价格预言机
		p.Price0CumulativeLast += _reserve1 / _reserve0 * uint64(timeElapsed)
		p.Price1CumulativeLast += _reserve0 / _reserve1 * uint64(timeElapsed)
	}
	p.BlockTimestampLast = blockTimestamp

	p.Reserve0 = balance0
	p.Reserve1 = balance1
}

func (p *Pair) UpdateKLast() {
	p.KLast = big.NewInt(0).Mul(big.NewInt(int64(p.Reserve0)), big.NewInt(int64(p.Reserve1)))
}

func DecodeToPair(bytes []byte) (*Pair, error) {
	var pair *Pair
	err := rlp.DecodeBytes(bytes, &pair)
	return pair, err
}
