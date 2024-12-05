package ssz_types

import (
	"bytes"
	"encoding/hex"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types/big"
)

func sampleFile() *SSZFile_v1 {
	out := NewSSZFile_v1()

	out.RewardsFileVersion = 10
	out.RulesetVersion = 4
	out.Network = 17000
	out.Index = 11
	out.StartTime = time.Now().Add(time.Hour * -24)
	out.EndTime = time.Now()
	out.ConsensusStartBlock = 128
	out.ConsensusEndBlock = 256
	out.ExecutionStartBlock = 1024
	out.ExecutionEndBlock = 1280
	out.IntervalsPassed = 1
	_, _ = hex.Decode(out.MerkleRoot[:], []byte("ac9ddbc55a8cd92612b86866de955f0bb99dd51e1447767afc610b13a5063546"))
	out.TotalRewards = &TotalRewards{
		ProtocolDaoRpl:               big.NewUint256(1000),
		TotalCollateralRpl:           big.NewUint256(2000),
		TotalOracleDaoRpl:            big.NewUint256(3000),
		TotalSmoothingPoolEth:        big.NewUint256(4000),
		PoolStakerSmoothingPoolEth:   big.NewUint256(5000),
		NodeOperatorSmoothingPoolEth: big.NewUint256(6000),
		TotalNodeWeight:              big.NewUint256(7000),
	}
	out.NetworkRewards = NetworkRewards{
		&NetworkReward{
			Network:          0,
			CollateralRpl:    big.NewUint256(200),
			OracleDaoRpl:     big.NewUint256(300),
			SmoothingPoolEth: big.NewUint256(400),
		},
		&NetworkReward{
			Network:          1,
			CollateralRpl:    big.NewUint256(500),
			OracleDaoRpl:     big.NewUint256(600),
			SmoothingPoolEth: big.NewUint256(700),
		},
	}
	out.NodeRewards = NodeRewards{
		&NodeReward{
			Address:          Address{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
			Network:          0,
			CollateralRpl:    big.NewUint256(10),
			OracleDaoRpl:     big.NewUint256(20),
			SmoothingPoolEth: big.NewUint256(30),
		},
		&NodeReward{
			Address:          Address{0x01, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
			Network:          1,
			CollateralRpl:    big.NewUint256(10),
			OracleDaoRpl:     big.NewUint256(20),
			SmoothingPoolEth: big.NewUint256(30),
		},
	}

	return out
}

func fatalIf(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatal(err)
}

func TestSSZFileRoundTrip(t *testing.T) {
	f := sampleFile()
	hashRoot, err := f.HashTreeRoot()
	t.Logf("Original hash root: %x", hashRoot)
	fatalIf(t, err)

	data, err := f.FinalizeSSZ()
	fatalIf(t, err)

	f, err = ParseSSZFile(data)
	fatalIf(t, err)
	hashRoot2, err := f.HashTreeRoot()
	t.Logf("Rount-trip hash root: %x", hashRoot2)
	fatalIf(t, err)

	if !bytes.Equal(hashRoot2[:], hashRoot[:]) {
		t.Fatal("Round-trip ssz differed from original ssz")
	}
}

func TestSSZFileJSONRoundTrip(t *testing.T) {
	f := sampleFile()
	hashRoot, err := f.HashTreeRoot()
	t.Logf("Original hash root: %x", hashRoot)
	fatalIf(t, err)

	data, err := f.MarshalJSON()
	fatalIf(t, err)

	f = &SSZFile_v1{}
	fatalIf(t, f.UnmarshalJSON(data))

	hashRoot2, err := f.HashTreeRoot()
	t.Logf("Rount-trip hash root: %x", hashRoot2)
	fatalIf(t, err)

	if !bytes.Equal(hashRoot2[:], hashRoot[:]) {
		t.Fatal("Round-trip ssz differed from original ssz")
	}
}

func TestSSZFileDuplicateNodeRewards(t *testing.T) {
	f := sampleFile()
	f.NodeRewards = append(f.NodeRewards, f.NodeRewards[1])
	err := f.Verify()
	if err == nil {
		t.Fatal("expected error due to duplicate entries")
	}
	if !strings.Contains(err.Error(), "duplicate entries") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileDuplicateNetworkRewards(t *testing.T) {
	f := sampleFile()
	f.NetworkRewards = append(f.NetworkRewards, f.NetworkRewards[1])
	err := f.Verify()
	if err == nil {
		t.Fatal("expected error due to duplicate entries")
	}
	if !strings.Contains(err.Error(), "duplicate entries") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileOutOfOrderNodeRewards(t *testing.T) {
	f := sampleFile()
	slices.Reverse(f.NodeRewards)
	err := f.Verify()
	if err == nil {
		t.Fatal("expected error due to sorting")
	}
	if !strings.Contains(err.Error(), "out of order") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileOutOfOrderNetworkRewards(t *testing.T) {
	f := sampleFile()
	slices.Reverse(f.NetworkRewards)
	err := f.Verify()
	if err == nil {
		t.Fatal("expected error due to sorting")
	}
	if !strings.Contains(err.Error(), "out of order") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileMissingTotalRewards(t *testing.T) {
	f := sampleFile()
	f.TotalRewards = nil
	err := f.Verify()
	if err == nil {
		t.Fatal("expected error due to missing field")
	}
	if !strings.Contains(err.Error(), "missing required field TotalRewards") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileUnknownNetwork(t *testing.T) {
	f := sampleFile()
	f.Network = 3
	hashRoot, err := f.HashTreeRoot()
	t.Logf("Original hash root: %x", hashRoot)
	fatalIf(t, err)

	data, err := f.MarshalJSON()
	fatalIf(t, err)

	f = &SSZFile_v1{}
	fatalIf(t, f.UnmarshalJSON(data))

	hashRoot2, err := f.HashTreeRoot()
	t.Logf("Rount-trip hash root: %x", hashRoot2)
	fatalIf(t, err)

	if !bytes.Equal(hashRoot2[:], hashRoot[:]) {
		t.Fatal("Round-trip ssz differed from original ssz")
	}
}

func TestSSZFileNoMagic(t *testing.T) {
	f := sampleFile()
	copy(f.Magic[:], []byte{0x00, 0x01, 0x02, 0x03})
	data, err := f.MarshalSSZ()
	fatalIf(t, err)
	f, err = ParseSSZFile(data)
	if err == nil {
		t.Fatal("expected error due to missing magic header")
	}
	if !strings.Contains(err.Error(), "magic header not found") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileBadRoot(t *testing.T) {
	f := sampleFile()
	copy(f.MerkleRoot[:], []byte{0x00, 0x01, 0x02, 0x03})
	data, err := f.MarshalSSZ()
	fatalIf(t, err)
	f, err = ParseSSZFile(data)
	if err == nil {
		t.Fatal("expected error due to mangled MerkleRoot")
	}
	if !strings.Contains(err.Error(), "mismatch against existing root") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileCalculateRoot(t *testing.T) {
	f := sampleFile()
	_, _ = hex.Decode(f.MerkleRoot[:], []byte("0000000000000000000000000000000000000000000000000000000000000000"))
	data, err := f.MarshalSSZ()
	fatalIf(t, err)
	f, err = ParseSSZFile(data)
	fatalIf(t, err)

	// Make sure the root is now set
	if bytes.Count(f.MerkleRoot[:], []byte{0x00}) >= 32 {
		t.Fatal("Expected ParseSSZFile to set the missing root")
	}
}

func TestSSZFileFinalizeFail(t *testing.T) {
	f := sampleFile()
	copy(f.MerkleRoot[:], []byte{0x00, 0x01, 0x02, 0x03})
	_, err := f.FinalizeSSZ()
	if err == nil {
		t.Fatal("expected error due to mangled MerkleRoot")
	}
	if !strings.Contains(err.Error(), "mismatch against existing root") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileTruncatedError(t *testing.T) {
	f := sampleFile()
	data, err := f.FinalizeSSZ()
	data = data[:10]
	f, err = ParseSSZFile(data)
	if err == nil {
		t.Fatal("expected error due to mangled file bytes")
	}
	if !strings.Contains(err.Error(), "incorrect size") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestSSZFileSorting(t *testing.T) {
	f := sampleFile()
	slices.Reverse(f.NetworkRewards)
	slices.Reverse(f.NodeRewards)
	sort.Sort(f.NetworkRewards)
	if !sort.IsSorted(f.NetworkRewards) {
		t.Fatal("sorting NetworkRewards failed")
	}
	sort.Sort(f.NodeRewards)
	if !sort.IsSorted(f.NodeRewards) {
		t.Fatal("sorting NodeRewards failed")
	}

}
