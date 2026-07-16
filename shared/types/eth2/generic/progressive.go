package generic

import "math/bits"

// Progressive Merkle tree helpers for EIP-7495 ProgressiveContainer and
// EIP-7916 ProgressiveList / ProgressiveBitlist (adopted by EIP-7688 in Gloas).
//
// These replace classic Container/List g-indices only for types that the Gloas
// specs redefine as progressive. Pre-Gloas forks (Deneb/Electra/Fulu) must keep
// using the classic helpers in gindex.go (ContainerFieldGindex,
// GetGeneralizedIndexForValidator, etc.).
//
// Progressive tree shape (chunk capacities grow 1, 4, 16, 64, … on successive
// levels; each level is a binary subtree on the left with the next progressive
// level on the right):
//
//	root
//	 /\
//	/  \
//
// L0  next0   // L0: chunks[0 .. 1)
//
//	/\
//
// L1  next1   // L1: chunks[1 .. 5)
//
// ProgressiveContainer root = hash(progressive_tree, active_fields)
// ProgressiveList root      = hash(progressive_tree, length)

// ConcatGindices concatenates generalized-index path steps the way remerkleable
// does: each step is a gindex relative to its own subtree root (root = 1).
func ConcatGindices(steps ...uint64) uint64 {
	out := uint64(1)
	for _, step := range steps {
		if step < 1 {
			// Invalid gindex; treat as no-op path to avoid panics. Callers must
			// only pass valid gindices (>= 1).
			continue
		}
		// Number of path bits after the leading 1.
		stepBitLen := bits.Len64(step) - 1
		out <<= stepBitLen
		out |= step ^ (uint64(1) << stepBitLen)
	}
	return out
}

// ProgressiveChunkGindex returns the generalized index of chunk `index` within
// a progressive Merkle tree whose root is gindex 1 (EIP-7916).
func ProgressiveChunkGindex(index uint64) uint64 {
	gindex := uint64(1)
	start := uint64(0)
	capacity := uint64(1)
	for {
		if index < start+capacity {
			offset := index - start
			// Left into this level's binary subtree, then to the leaf.
			gindex = gindex * 2
			gindex = gindex*capacity + offset
			return gindex
		}
		// Right into the next progressive level.
		gindex = gindex*2 + 1
		start += capacity
		capacity *= 4
	}
}

// ProgressiveContainerFieldGindex returns the generalized index of field
// `fieldIndex` (its ssz-index) within a ProgressiveContainer whose root is
// gindex 1. The progressive tree of field roots is the left child; active_fields
// is mixed in on the right.
func ProgressiveContainerFieldGindex(fieldIndex uint64) uint64 {
	return ConcatGindices(2, ProgressiveChunkGindex(fieldIndex))
}

// ProgressiveListElementGindex returns the generalized index of element
// `elementIndex` within a ProgressiveList whose root is gindex 1. The progressive
// tree of elements is the left child; the list length is mixed in on the right.
//
// This is the index of the Merkle chunk for that element. For ProgressiveList of
// composite types (e.g. Validator), each element is one chunk so elementIndex is
// used directly. For basic types that pack into 32-byte chunks (e.g. uint64 packs
// 4 per leaf), pass the packed chunk index instead of the raw element index.
func ProgressiveListElementGindex(elementIndex uint64) uint64 {
	return ConcatGindices(2, ProgressiveChunkGindex(elementIndex))
}

// GetGeneralizedIndexForProgressiveListElement returns the gindex of an element
// inside a ProgressiveList that itself sits at listRootGindex in a larger tree
// (e.g. a ProgressiveContainer field). See ProgressiveListElementGindex for
// packing notes on basic vs composite element types.
//
// Gloas only for ProgressiveList fields. Fixed-capacity List navigation (including
// Gloas historical_summaries) uses GetGeneralizedIndexForListElement in gindex.go.
func GetGeneralizedIndexForProgressiveListElement(listRootGindex uint64, elementIndex uint64) uint64 {
	return ConcatGindices(listRootGindex, ProgressiveListElementGindex(elementIndex))
}
