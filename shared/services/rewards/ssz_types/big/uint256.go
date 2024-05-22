package big

import (
	"math/big"

	ssz "github.com/ferranbt/fastssz"
	"github.com/holiman/uint256"
)

type Uint256 struct {
	repr *uint256.Int `ssz:"-"`
}

func NewUint256(i uint64) Uint256 {
	out := Uint256{}
	out.repr = uint256.NewInt(i)
	return out
}

func (u *Uint256) Unwrap() *uint256.Int {
	out := uint256.NewInt(0)
	out.Set(u.repr)
	return out
}

func (u *Uint256) SizeSSZ() (size int) {
	return 32
}

func (u *Uint256) MarshalSSZTo(buf []byte) ([]byte, error) {
	bytes, err := u.repr.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return append(buf, bytes...), nil
}

func (u *Uint256) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	bytes := make([]byte, 32)
	bytes, err = u.MarshalSSZTo(bytes)
	if err != nil {
		return
	}

	hh.AppendBytes32(bytes)
	return
}

func (u *Uint256) UnmarshalSSZ(buf []byte) error {
	u.repr = uint256.NewInt(0)
	return u.repr.UnmarshalSSZ(buf)
}

func (u *Uint256) String() string {
	return u.repr.String()
}

func (u *Uint256) UnmarshalJSON(data []byte) error {
	u.repr = uint256.NewInt(0)
	return u.repr.UnmarshalJSON(data)
}

func (u Uint256) MarshalJSON() ([]byte, error) {
	return u.repr.MarshalJSON()
}

func (u Uint256) ToBig() *big.Int {
	return u.repr.ToBig()
}
