package rewards

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Reads an existing RewardsFile from disk and wraps it in a LocalFile
func ReadLocalRewardsFile(path string) (*LocalRewardsFile, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading rewards file from %s: %w", path, err)
	}

	// Unmarshal it
	proofWrapper, err := DeserializeRewardsFile(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling rewards file from %s: %w", path, err)
	}

	return NewLocalFile[IRewardsFile](proofWrapper, path), nil
}

// Reads an existing MinipoolPerformanceFile from disk and wraps it in a LocalFile
func ReadLocalMinipoolPerformanceFile(path string) (*LocalMinipoolPerformanceFile, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading rewards file from %s: %w", path, err)
	}

	// Unmarshal it
	minipoolPerformance, err := DeserializeMinipoolPerformanceFile(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling rewards file from %s: %w", path, err)
	}

	return NewLocalFile[IMinipoolPerformanceFile](minipoolPerformance, path), nil
}

// Interface for local rewards or minipool performance files
type ISerializable interface {
	// Converts the underlying interface to a byte slice
	Serialize() ([]byte, error)
}

// A wrapper around ISerializable representing a local rewards file or minipool performance file.
// Can be used with anything that can be serialzed to bytes or parsed from bytes.
type LocalFile[T ISerializable] struct {
	f        T
	fullPath string
}

type ILocalFile interface {
	ISerializable
	Write() ([]byte, error)
	Path() string
	FileName() string
	CreateCompressedFileAndCid() (string, cid.Cid, error)
}

// Type aliases
type LocalRewardsFile = LocalFile[IRewardsFile]
type LocalMinipoolPerformanceFile = LocalFile[IMinipoolPerformanceFile]

// NewLocalFile creates the wrapper, but doesn't write to disk.
// This should be used when generating new trees / performance files.
func NewLocalFile[T ISerializable](ilf T, fullpath string) *LocalFile[T] {
	return &LocalFile[T]{
		f:        ilf,
		fullPath: fullpath,
	}
}

// Returns the underlying interface, IRewardsFile for rewards file, IMinipoolPerformanceFile for performance, etc.
func (lf *LocalFile[T]) Impl() T {
	return lf.f
}

// Converts the underlying interface to a byte slice
func (lf *LocalFile[T]) Serialize() ([]byte, error) {
	return lf.f.Serialize()
}

// Serializes the file and writes it to disk
func (lf *LocalFile[T]) Write() ([]byte, error) {
	data, err := lf.Serialize()
	if err != nil {
		return nil, fmt.Errorf("error serializing file: %w", err)
	}

	err = os.WriteFile(lf.fullPath, data, 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing file to %s: %w", lf.fullPath, err)
	}
	return data, nil
}

func (lf *LocalFile[T]) Path() string {
	return lf.fullPath
}

func (lf *LocalFile[T]) FileName() string {
	return filepath.Base(lf.Path())
}

// Computes the CID that would be used if we compressed the file with zst,
// added the ipfs extension to the filename (.zst), and uploaded it to ipfs
// in an empty directory, as web3storage did, once upon a time.
//
// N.B. This function will also save the compressed file to disk so it can
// later be uploaded to ipfs
func (lf *LocalFile[T]) CreateCompressedFileAndCid() (string, cid.Cid, error) {
	// Serialize
	data, err := lf.Serialize()
	if err != nil {
		return "", cid.Cid{}, fmt.Errorf("error serializing file: %w", err)
	}

	// Compress
	encoder, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	compressedBytes := encoder.EncodeAll(data, make([]byte, 0, len(data)))

	filename := lf.fullPath + config.RewardsTreeIpfsExtension
	c, err := singleFileDirIPFSCid(compressedBytes, filepath.Base(filename))
	if err != nil {
		return filename, cid.Cid{}, fmt.Errorf("error calculating CID: %w", err)
	}

	// Write to disk
	// Take care to write to `filename` since it has the .zst extension added
	err = os.WriteFile(filename, compressedBytes, 0644)
	if err != nil {
		return filename, cid.Cid{}, fmt.Errorf("error writing file to %s: %w", lf.fullPath, err)
	}
	return filename, c, nil
}

// Saves all rewards artifacts, including ssz if the rewards file is at least v3.
// If nodeTrusted is passed, zstd compressed copies will also be saved, with the cid of the
// compressed minipool perf file added to the rewards file before the latter is compressed.
//
// If the rewards file is at least v3, the cid of the uncompressed ssz file is returned for consensus
// Otherwise, the cid of the compressed json rewards file is returned for consensus.
// Thus, this function is only suitable for v9+ and versions below should use saveJSONArtifacts instead
func saveRewardsArtifacts(smartnode *config.SmartnodeConfig, rewardsFile IRewardsFile, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error) {
	if rewardsFile.GetHeader().RewardsFileVersion < rewardsFileVersionThree {
		return saveJSONArtifacts(smartnode, rewardsFile, nodeTrusted)
	}

	return saveArtifactsImpl(smartnode, rewardsFile, nodeTrusted, true)
}

// Saves JSON artifacts from tree generation
// If nodeTrusted is passed, zstd compressed copies will also be saved, with the cid of the
// compressed minipool perf file added to the rewards file before the latter is compressed.
//
// Returns the cid of the compressed rewards file, a map containing all the other cids, or an error.
func saveJSONArtifacts(smartnode *config.SmartnodeConfig, rewardsFile IRewardsFile, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error) {
	return saveArtifactsImpl(smartnode, rewardsFile, nodeTrusted, false)
}

// Saves JSON artifacts and optionally compressed + ssz artifacts
// If includeSSZ is true, the primary cid is the uncompressed reward ssz.
// Otherwise, it is the compressed rewards json.
func saveArtifactsImpl(smartnode *config.SmartnodeConfig, rewardsFile IRewardsFile, nodeTrusted bool, includeSSZ bool) (cid.Cid, map[string]cid.Cid, error) {
	currentIndex := rewardsFile.GetHeader().Index

	var primaryCid *cid.Cid
	out := make(map[string]cid.Cid, 4)

	for i, f := range []ILocalFile{
		// Do not reorder!
		// i == 0 - minipool performance file
		NewLocalFile[IMinipoolPerformanceFile](
			rewardsFile.GetMinipoolPerformanceFile(),
			smartnode.GetMinipoolPerformancePath(currentIndex, true),
		),
		// i == 1 - rewards file
		NewLocalFile[IRewardsFile](
			rewardsFile,
			smartnode.GetRewardsTreePath(currentIndex, true, config.RewardsExtensionJSON),
		),
		// i == 2 - ssz rewards file
		//NewLocalFile[IRewardsFile](
		//	rewardsFile.SSZFile(),
		//	smartnode.GetRewardsTreePath(currentIndex, true, config.RewardsExtensionSSZ),
		//),
	} {

		data, err := f.Write()
		if err != nil {
			return cid.Cid{}, nil, fmt.Errorf("error saving %s: %w", f.Path(), err)
		}

		uncompressedCid, err := singleFileDirIPFSCid(data, f.FileName())
		if err != nil {
			return cid.Cid{}, nil, fmt.Errorf("error calculating cid for saved file %s: %w", f.Path(), err)
		}
		out[f.FileName()] = uncompressedCid

		if !nodeTrusted {
			// For some reason we didn't simply omit this in the past, so for consistency, keep setting it.
			rewardsFile.SetMinipoolPerformanceFileCID("---")
			// Non odao nodes only need inflated files
			continue
		}

		// Save compressed versions
		compressedFilePath, compressedCid, err := f.CreateCompressedFileAndCid()
		if err != nil {
			return cid.Cid{}, nil, fmt.Errorf("error compressing file %s: %w", f.Path(), err)
		}
		out[filepath.Base(compressedFilePath)] = compressedCid

		// Note the performance cid in the rewards file
		if i == 0 {
			rewardsFile.SetMinipoolPerformanceFileCID(compressedCid.String())
		}

		// Note the primary cid for artifacts used for consensus
		if !includeSSZ {
			// JSON rewards file
			if i == 1 {
				primaryCid = &compressedCid
			}
		} else {
			// SSZ rewards file
			if i == 2 {
				// Consensus is on the uncompressed cid when using ssz
				primaryCid = &uncompressedCid
			}
		}

	}
	return *primaryCid, out, nil
}
