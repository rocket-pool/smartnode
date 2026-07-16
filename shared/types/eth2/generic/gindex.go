package generic

import (
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Pre-Gloas SSZ generalized-index helpers.
//
// Use these for Deneb / Electra / Fulu where it still merkleizes
// BeaconState / lists with pre-Gloas SSZ Container and List[T, N] trees.
//
// Gloas (EIP-7688) switched evolving containers/lists to ProgressiveContainer /
// ProgressiveList. For those types use the helpers in progressive.go instead:
// ProgressiveContainerFieldGindex, ProgressiveListElementGindex, etc.
//
// Nested Vector and fixed-capacity List navigation under either style of parent
// still uses GetGeneralizedIndexForVectorElement / GetGeneralizedIndexForListElement.

// ContainerFieldGindex returns the generalized index of field `fieldIndex`
// inside a standard SSZ Container that has `fieldCount` fields.
//
// Fields are left-aligned in a binary tree of width nextPowerOfTwo(fieldCount),
// so the gindex is nextPowerOfTwo(fieldCount) + fieldIndex (root = 1).
//
// Pre-Gloas only. Progressive containers use ProgressiveContainerFieldGindex.
func ContainerFieldGindex(fieldCount, fieldIndex uint64) uint64 {
	return math.GetPowerOfTwoCeil(fieldCount) + fieldIndex
}

// GetGeneralizedIndexForValidator returns the gindex of validators[index] when
// validators is a fixed-capacity SSZ List[Validator, VALIDATOR_REGISTRY_LIMIT]
// (capacity 2^40) sitting at validatorsArrayIndex in the parent tree.
//
// Pre-Gloas only. Gloas validators are a ProgressiveList; use
// GetGeneralizedIndexForProgressiveListElement (or gloas.GetGeneralizedIndexForValidator).
func GetGeneralizedIndexForValidator(index uint64, validatorsArrayIndex uint64) uint64 {
	// List root mixes in length: left = data tree, right = length.
	// Composite list elements are one leaf each; capacity is power-of-two max length.
	return GetGeneralizedIndexForListElement(validatorsArrayIndex, beaconStateValidatorsMaxLength, index)
}

// GetGeneralizedIndexForListElement returns the gindex of an element inside a
// fixed-capacity SSZ List[T, N] (with length mixin) that sits at listRootGindex.
// capacity must be a power of two (typically the list's N).
//
// Applies under both classic Containers and ProgressiveContainers once you have
// the list field's root gindex (historical_summaries stays a normal List in Gloas).
func GetGeneralizedIndexForListElement(listRootGindex uint64, capacity uint64, elementIndex uint64) uint64 {
	// List root: left = data tree, right = length. Data leaves start at
	// listRoot * 2 * capacity + elementIndex.
	return listRootGindex*2*capacity + elementIndex
}

// GetGeneralizedIndexForVectorElement returns the gindex of an element inside a
// fixed-length SSZ Vector[T, N] that sits at vectorRootGindex. length must be a
// power of two.
//
// Applies under both classic Containers and ProgressiveContainers once you have
// the vector field's root gindex (e.g. block_roots).
func GetGeneralizedIndexForVectorElement(vectorRootGindex uint64, length uint64, elementIndex uint64) uint64 {
	return vectorRootGindex*length + elementIndex
}
