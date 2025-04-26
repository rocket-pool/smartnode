package eth2

// BeaconStateValidatorsIndex is the field offset of the Validators field in the BeaconState struct
const beaconStateValidatorsIndex uint64 = 11

// If this ever isn't a power of two, we need to round up to the next power of two
const beaconStateValidatorsMaxLength uint64 = 1 << 40

const beaconStateHistoricalSummariesFieldIndex uint64 = 27
const beaconStateHistoricalSummariesMaxLength uint64 = 1 << 24
const beaconStateHistoricalSummaryChunkCeil uint64 = 2
const beaconStateHistoricalSummaryBlockSummaryRootIndex uint64 = 0
const beaconStateBlockRootsMaxLength uint64 = 1 << 13
const beaconStateBlockRootsFieldIndex uint64 = 5
