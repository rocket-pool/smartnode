package rewards

import (
	"os"
	"path"
	"testing"

	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
)

func TestFilesFromTree(t *testing.T) {
	dir := t.TempDir()
	t.Logf("%s using tempdir %s\n", t.Name(), dir)

	f := RewardsFile_v3{
		RewardsFileHeader: &sharedtypes.RewardsFileHeader{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
		MinipoolPerformanceFile: MinipoolPerformanceFile_v3{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
	}

	localRewardsFile := NewLocalFile[sharedtypes.IRewardsFile](
		&f,
		path.Join(dir, "rewards.json"),
	)

	err := localRewardsFile.Write()
	if err != nil {
		t.Fatal(err)
	}

	minipoolPerformanceFile := localRewardsFile.Impl().GetMinipoolPerformanceFile()
	localMinipoolPerformanceFile := NewLocalFile[sharedtypes.IMinipoolPerformanceFile](
		minipoolPerformanceFile,
		path.Join(dir, "performance.json"),
	)

	err = localMinipoolPerformanceFile.Write()
	if err != nil {
		t.Fatal(err)
	}

	// Check that the file can be parsed
	localRewardsFile, err = ReadLocalRewardsFile(path.Join(dir, "rewards.json"))
	if err != nil {
		t.Fatal(err)
	}

	if localRewardsFile.Impl().GetHeader().RulesetVersion != f.RewardsFileHeader.RulesetVersion {
		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			localRewardsFile.Impl().GetHeader().RulesetVersion,
			f.RewardsFileHeader.RulesetVersion,
		)
	}

	localMinipoolPerformanceFile, err = ReadLocalMinipoolPerformanceFile(path.Join(dir, "performance.json"))
	if err != nil {
		t.Fatal(err)
	}

}

func TestCompressionAndCids(t *testing.T) {
	dir := t.TempDir()
	t.Logf("%s using tempdir %s\n", t.Name(), dir)

	f := RewardsFile_v3{
		RewardsFileHeader: &sharedtypes.RewardsFileHeader{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
		MinipoolPerformanceFile: MinipoolPerformanceFile_v3{
			RewardsFileVersion: 3,
			RulesetVersion:     9,
		},
	}

	localRewardsFile := NewLocalFile[sharedtypes.IRewardsFile](
		&f,
		path.Join(dir, "rewards.json"),
	)

	minipoolPerformanceFile := localRewardsFile.Impl().GetMinipoolPerformanceFile()
	localMinipoolPerformanceFile := NewLocalFile[sharedtypes.IMinipoolPerformanceFile](
		minipoolPerformanceFile,
		path.Join(dir, "performance.json"),
	)

	rewardsCid, err := localRewardsFile.CreateCompressedFileAndCid()
	if err != nil {
		t.Fatal(err)
	}

	performanceCid, err := localMinipoolPerformanceFile.CreateCompressedFileAndCid()
	if err != nil {
		t.Fatal(err)
	}

	// Check that compressed files were written to disk and their cids match what was returned by CompressedCid
	compressedRewardsBytes, err := os.ReadFile(path.Join(dir, "rewards.json.zst"))
	if err != nil {
		t.Fatal(err)
	}

	rewardsFileCid, err := singleFileDirIPFSCid(compressedRewardsBytes, "rewards.json.zst")
	if err != nil {
		t.Fatal(err)
	}
	if rewardsFileCid != rewardsCid {
		t.Fatalf("expected CompressedCid to return %s, got %s", rewardsFileCid, rewardsCid)
	}

	compressedPerformanceBytes, err := os.ReadFile(path.Join(dir, "performance.json.zst"))
	if err != nil {
		t.Fatal(err)
	}

	performanceFileCid, err := singleFileDirIPFSCid(compressedPerformanceBytes, "performance.json.zst")
	if err != nil {
		t.Fatal(err)
	}
	if performanceFileCid != performanceCid {
		t.Fatalf("expected CompressedCid to return %s, got %s", performanceFileCid, performanceCid)
	}

	// Ensure that we can decompress both files
	decompressedPerformance, err := decompressFile(compressedPerformanceBytes)
	if err != nil {
		t.Fatal(err)
	}

	decompressedRewards, err := decompressFile(compressedRewardsBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure that we can parse the result of decompressing
	parsedPerformance, err := DeserializeMinipoolPerformanceFile(decompressedPerformance)
	if err != nil {
		t.Fatal(err)
	}

	parsedRewards, err := DeserializeRewardsFile(decompressedRewards)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure values were preserved in the round trip
	if localRewardsFile.Impl().GetHeader().RulesetVersion != parsedRewards.GetHeader().RulesetVersion {
		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			localRewardsFile.Impl().GetHeader().RulesetVersion,
			parsedRewards.GetHeader().RulesetVersion,
		)
	}

	if localRewardsFile.Impl().GetMinipoolPerformanceFile().(*MinipoolPerformanceFile_v3).RulesetVersion !=
		parsedPerformance.(*MinipoolPerformanceFile_v3).RulesetVersion {

		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			localRewardsFile.Impl().GetMinipoolPerformanceFile().(*MinipoolPerformanceFile_v3).RulesetVersion,
			parsedPerformance.(*MinipoolPerformanceFile_v3).RulesetVersion,
		)
	}
}
