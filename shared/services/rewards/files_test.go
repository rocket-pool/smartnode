package rewards

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestFilesFromTree(t *testing.T) {
	dir := t.TempDir()
	t.Logf("%s using tempdir %s\n", t.Name(), dir)

	f := RewardsFile_v3{
		RewardsFileHeader: &RewardsFileHeader{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
		MinipoolPerformanceFile: MinipoolPerformanceFile_v2{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
	}

	localRewardsFile := NewLocalFile[IRewardsFile](
		&f,
		path.Join(dir, "rewards.json"),
	)

	rewardsFileBytes, err := localRewardsFile.Write()
	if err != nil {
		t.Fatal(err)
	}
	if rewardsFileBytes == nil {
		t.Fatal("Write() should have returned serialized data")
	}
	directBytes, _ := f.Serialize()
	if !bytes.Equal(directBytes, rewardsFileBytes) {
		t.Fatal("Write() returned something different than Serialize()")
	}

	minipoolPerformanceFile := &f.MinipoolPerformanceFile
	localMinipoolPerformanceFile := NewLocalFile[IPerformanceFile](
		minipoolPerformanceFile,
		path.Join(dir, "performance.json"),
	)

	miniPerfFileBytes, err := localMinipoolPerformanceFile.Write()
	if err != nil {
		t.Fatal(err)
	}
	if miniPerfFileBytes == nil {
		t.Fatal("Write() should have returned serialized data")
	}
	directBytes, _ = minipoolPerformanceFile.Serialize()
	if !bytes.Equal(directBytes, miniPerfFileBytes) {
		t.Fatal("Write() returned something different than Serialize()")
	}

	// Check that the file can be parsed
	localRewardsFile, err = ReadLocalRewardsFile(path.Join(dir, "rewards.json"))
	if err != nil {
		t.Fatal(err)
	}

	if localRewardsFile.Impl().(*RewardsFile_v3).RulesetVersion != f.RewardsFileHeader.RulesetVersion {
		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			localRewardsFile.Impl().(*RewardsFile_v3).RulesetVersion,
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
		RewardsFileHeader: &RewardsFileHeader{
			RewardsFileVersion: 3,
			RulesetVersion:     8,
		},
		MinipoolPerformanceFile: MinipoolPerformanceFile_v2{
			RewardsFileVersion: 3,
			RulesetVersion:     9,
		},
	}

	localRewardsFile := NewLocalFile[IRewardsFile](
		&f,
		path.Join(dir, "rewards.json"),
	)

	minipoolPerformanceFile := &f.MinipoolPerformanceFile
	localMinipoolPerformanceFile := NewLocalFile[IPerformanceFile](
		minipoolPerformanceFile,
		path.Join(dir, "performance.json"),
	)

	returnedFilename, rewardsCid, err := localRewardsFile.CreateCompressedFileAndCid()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(returnedFilename) != "rewards.json.zst" {
		t.Fatalf("Unexpected filename: %s", returnedFilename)
	}

	returnedFilename, performanceCid, err := localMinipoolPerformanceFile.CreateCompressedFileAndCid()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(returnedFilename) != "performance.json.zst" {
		t.Fatalf("Unexpected filename: %s", returnedFilename)
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
	if localRewardsFile.Impl().(*RewardsFile_v3).RulesetVersion != parsedRewards.(*RewardsFile_v3).RulesetVersion {
		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			localRewardsFile.Impl().(*RewardsFile_v3).RulesetVersion,
			parsedRewards.(*RewardsFile_v3).RulesetVersion,
		)
	}

	if minipoolPerformanceFile.RulesetVersion !=
		parsedPerformance.(*MinipoolPerformanceFile_v2).RulesetVersion {

		t.Fatalf(
			"expected parsed version %d to match serialized version %d\n",
			minipoolPerformanceFile.RulesetVersion,
			parsedPerformance.(*MinipoolPerformanceFile_v2).RulesetVersion,
		)
	}
}

// Methodology:
// First, we test against rp-rewards-mainnet-17.json, the official version from ipfs.
// Once manually confirming the CID matches the on-chain value, we test against a smaller
// file.
//
// The CID in this test _should never be updated_.
//
// If new code changes the ipfs cid calculation, it needs to do so in a way that preserves
// the pre-existing behavior for historical trees.
func TestCidConsistency(t *testing.T) {
	/* Commenting out the code used to verify the cid calculator matched an on-chain cid
	// Load rp-rewards-mainnet-17.json and check the resulting cid
	localRewardsFile, err := ReadLocalRewardsFile("rp-rewards-mainnet-17.json")
	if err != nil {
		t.Fatal(err)
	}

	// Validate its CID against the on-chain version
	cid, err := localRewardsFile.CreateCompressedFileAndCid()
	if err != nil {
		t.Fatal(err)
	}

	err = localRewardsFile.Write()
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(cid) != "bafybeifldymulw6qvlfjgntj6mrlbwcl46xn6njickcydlquxc33nseoxi" {
		t.Fatalf("unexpected cid %s", cid)
	}
	*/

	// 256 random bytes from /dev/urandom on a cloudy day in Brooklyn
	// dd if=/dev/urandom bs=1 count=256 | xxd -p
	randomCharacters := `e73d6923a8b99cbd9de59619626292b5173f27ddaf50f21ce885272ab63060c8acfe10a066b24a457232afa00ef23f8be61d112935dbaa81658ba1699e5eef9dd973ac2c8d7ecbaee7063c25ca040eb446139cf99630510b3514ff5c4c2d5be13a2a73cb55cf27e743b2f317153fbbfd3f8e3c3c788160a2458c69c6fd905fd4ce5afc3634532d1f6e2e27fb1cb049356d8ccc6599710d82cf75b65f2d03e6d969d0200b18f0217e3aa500a5053636f105126ff0d00c6b8e0f47f2cc5f1ec73bc9e66f023f79ab09fd3a5f7c5ee988ec4028479026bc02fb1ab22f50eaf985c1d0c357cdeca0cfbe49e465fb3967a42b4d2e63949910cef8487ba5853eaee442`
	data, err := hex.DecodeString(randomCharacters)
	if err != nil {
		t.Fatal(err)
	}

	cid, err := singleFileDirIPFSCid(data, "test.bin")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Computed CID: %s", fmt.Sprint(cid))
	if "bafybeibqxb2xeoh2mlcn7543jr3tgvdu74mqqd43esrttyktmu3ubtx63i" != fmt.Sprint(cid) {
		t.Fatal("CID did not match expectations. If changing CID computation logic, ensure historical CIDs can be recomputed. See comments in files_test.go for more info")
	}
}
