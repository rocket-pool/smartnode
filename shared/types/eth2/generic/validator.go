package generic

import (
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func GetGeneralizedIndexForValidator(index uint64, validatorsArrayIndex uint64) uint64 {
	root := validatorsArrayIndex

	// Now, grab the validator index within the list
	// `start` is `index * 32` and `pos` is `start / 32` so pos is just `index`
	pos := index
	baseIndex := uint64(2) // Lists have a base index of 2
	root = root*baseIndex*math.GetPowerOfTwoCeil(beaconStateValidatorsMaxLength) + pos

	// root is now the generalized index for the validator
	return root
}
