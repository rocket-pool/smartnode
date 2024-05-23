package big

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	ssz "github.com/ferranbt/fastssz"
	"github.com/holiman/uint256"
)

var Overflow = errors.New("uint256 overflow")
var Negative = errors.New("uint256 can't be negative before serializing")

// Wraps big.Int but will be checked for sign/overflow when serializing SSZ
type Uint256 struct {
	*big.Int
}

func NewUint256(i int64) Uint256 {
	return Uint256{big.NewInt(i)}
}

func (u *Uint256) SizeSSZ() (size int) {
	return 32
}

func (u *Uint256) ToUint256() (*uint256.Int, error) {
	// Check sign
	if u.Sign() < 0 {
		return nil, Negative
	}

	s, overflow := uint256.FromBig(u.Int)
	if overflow {
		return nil, Overflow
	}
	return s, nil
}

func (u *Uint256) MarshalSSZTo(buf []byte) ([]byte, error) {
	s, err := u.ToUint256()
	if err != nil {
		return nil, err
	}

	bytes, err := s.MarshalSSZ()
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
	repr := uint256.NewInt(0)
	err := repr.UnmarshalSSZ(buf)
	if err != nil {
		return err
	}
	u.Int = repr.ToBig()
	return nil
}

func (u *Uint256) String() string {
	return u.Int.String()
}

func (u *Uint256) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	u.Int = big.NewInt(0)
	return u.Int.UnmarshalJSON([]byte(s))
}

func (u *Uint256) MarshalJSON() ([]byte, error) {
	s, err := u.Int.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\"%s\"", s)), nil
}

func (u *Uint256) Bytes32() ([32]byte, error) {
	s, err := u.ToUint256()
	if err != nil {
		return [32]byte{}, err
	}

	return s.Bytes32(), nil
}
