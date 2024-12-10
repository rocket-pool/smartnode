package ssz_types

import (
	"encoding/hex"
)

func (h Hash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}
