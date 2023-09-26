package types

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Information about an interval
type IntervalInfo struct {
	Index                  uint64        `json:"index"`
	TreeFilePath           string        `json:"treeFilePath"`
	TreeFileExists         bool          `json:"treeFileExists"`
	MerkleRootValid        bool          `json:"merkleRootValid"`
	CID                    string        `json:"cid"`
	StartTime              time.Time     `json:"startTime"`
	EndTime                time.Time     `json:"endTime"`
	NodeExists             bool          `json:"nodeExists"`
	CollateralRplAmount    *QuotedBigInt `json:"collateralRplAmount"`
	ODaoRplAmount          *QuotedBigInt `json:"oDaoRplAmount"`
	SmoothingPoolEthAmount *QuotedBigInt `json:"smoothingPoolEthAmount"`
	MerkleProof            []common.Hash `json:"merkleProof"`
}

type QuotedBigInt struct {
	big.Int
}

func NewQuotedBigInt(x int64) *QuotedBigInt {
	q := QuotedBigInt{}
	native := big.NewInt(x)
	q.Int = *native
	return &q
}

func (b *QuotedBigInt) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.String() + "\""), nil
}

func (b *QuotedBigInt) UnmarshalJSON(p []byte) error {
	strippedString := strings.Trim(string(p), "\"")
	nativeInt, success := big.NewInt(0).SetString(strippedString, 0)
	if !success {
		return fmt.Errorf("%s is not a valid big integer", strippedString)
	}

	b.Int = *nativeInt
	return nil
}
