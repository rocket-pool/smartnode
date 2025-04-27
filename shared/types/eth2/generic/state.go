package generic

// BeaconStateValidatorsIndex is the field offset of the Validators field in the BeaconState struct
const BeaconStateValidatorsIndex uint64 = 11

// If this ever isn't a power of two, we need to round up to the next power of two
const beaconStateValidatorsMaxLength uint64 = 1 << 40

const BeaconStateHistoricalSummariesFieldIndex uint64 = 27
const BeaconStateHistoricalSummariesMaxLength uint64 = 1 << 24
const beaconStateHistoricalSummaryChunkCeil uint64 = 2
const beaconStateHistoricalSummaryBlockSummaryRootIndex uint64 = 0
const BeaconStateBlockRootsMaxLength uint64 = 1 << 13
const BeaconStateBlockRootsFieldIndex uint64 = 5
