package generic

import (
	dynssz "github.com/pk910/dynamic-ssz"
)

// SSZ is the process-wide dynamic-ssz instance used for all beacon state and
// block encoding, hashing and merkle tree operations. The mainnet preset is
// baked into the static ssz-size/ssz-max struct tags, so no spec overrides are
// needed. Type descriptors are cached per instance — always use this singleton.
// The stream reader buffer is sized for beacon state downloads (~310 MB on
// mainnet); the 2 KB default would make far too many small reads.
var SSZ = dynssz.NewDynSsz(nil, dynssz.WithStreamReaderBufferSize(1<<20))
