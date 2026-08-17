package rewards

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const interval0YAML = `index: '0'
executionBlock: '15451078'
consensusBlock: '4598879'
merkleRoot: '0xb839fa0f5842bf3c8f19091361889fb0f1cb399d64b8da476d372b7de7a93463'
intervalsPassed: '1'
treasuryRPL: '10633670478560109530497'
trustedNodeRPL:
- '10633670478560109529794'
nodeRPL:
- '49623795566613844471758'
nodeETH:
- '0'
userETH: '0'
intervalStartTime: 1659591339
intervalEndTime: 1662010539
submissionTime: 1662011717
`

const interval44YAML = `index: '44'
executionBlock: '24238059'
consensusBlock: '13469279'
merkleRoot: '0x72ed38107b2b6c1a0469800af741c04adf5563aebe5a0ae5d77fb9b2354c64a5'
intervalsPassed: '1'
treasuryRPL: '22719124146910393305039'
trustedNodeRPL:
- '2065374922446399391296'
nodeRPL:
- '57830497828499182955594'
nodeETH:
- '15367476931278801422'
userETH: '4223474559072201135'
intervalStartTime: 1766036139
intervalEndTime: 1768455339
submissionTime: 1768461851
`

func TestParseRewardsEventFileInterval44(t *testing.T) {
	event, err := parseRewardsEventFile([]byte(interval44YAML), 44)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	assertBig(t, "index", event.Index, "44")
	assertBig(t, "executionBlock", event.ExecutionBlock, "24238059")
	assertBig(t, "consensusBlock", event.ConsensusBlock, "13469279")
	assertBig(t, "intervalsPassed", event.IntervalsPassed, "1")
	assertBig(t, "treasuryRPL", event.TreasuryRPL, "22719124146910393305039")
	assertBig(t, "userETH", event.UserETH, "4223474559072201135")
	if len(event.TrustedNodeRPL) != 1 {
		t.Fatalf("trustedNodeRPL len = %d", len(event.TrustedNodeRPL))
	}
	assertBig(t, "trustedNodeRPL[0]", event.TrustedNodeRPL[0], "2065374922446399391296")
	assertBig(t, "nodeRPL[0]", event.NodeRPL[0], "57830497828499182955594")
	assertBig(t, "nodeETH[0]", event.NodeETH[0], "15367476931278801422")
	wantRoot := common.HexToHash("0x72ed38107b2b6c1a0469800af741c04adf5563aebe5a0ae5d77fb9b2354c64a5")
	if event.MerkleRoot != wantRoot {
		t.Fatalf("merkleRoot = %s, want %s", event.MerkleRoot.Hex(), wantRoot.Hex())
	}
	if event.IntervalStartTime.Unix() != 1766036139 {
		t.Fatalf("intervalStartTime = %d", event.IntervalStartTime.Unix())
	}
	if event.IntervalEndTime.Unix() != 1768455339 {
		t.Fatalf("intervalEndTime = %d", event.IntervalEndTime.Unix())
	}
	if event.SubmissionTime.Unix() != 1768461851 {
		t.Fatalf("submissionTime = %d", event.SubmissionTime.Unix())
	}
}

func TestParseRewardsEventFileInterval0(t *testing.T) {
	event, err := parseRewardsEventFile([]byte(interval0YAML), 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	assertBig(t, "treasuryRPL", event.TreasuryRPL, "10633670478560109530497")
	assertBig(t, "userETH", event.UserETH, "0")
	assertBig(t, "nodeETH[0]", event.NodeETH[0], "0")
	if event.MerkleRoot != common.HexToHash("0xb839fa0f5842bf3c8f19091361889fb0f1cb399d64b8da476d372b7de7a93463") {
		t.Fatalf("merkleRoot = %s", event.MerkleRoot.Hex())
	}
}

func TestParseRewardsEventFileIndexMismatch(t *testing.T) {
	_, err := parseRewardsEventFile([]byte(interval44YAML), 43)
	if err == nil {
		t.Fatal("expected index mismatch error")
	}
}

func TestGetIntervalRewardsEventUnsupportedNetwork(t *testing.T) {
	event, ok, err := getIntervalRewardsEvent(44, "devnet", t.TempDir())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if ok {
		t.Fatalf("expected miss for devnet, got %+v", event)
	}
	_, ok, err = getIntervalRewardsEvent(44, "", t.TempDir())
	if err != nil || ok {
		t.Fatalf("expected miss for empty network, ok=%v err=%v", ok, err)
	}
}

func TestGetIntervalRewardsEventCacheHit(t *testing.T) {
	cacheDir := t.TempDir()
	path := rewardsEventLocalPath(cacheDir, networkMainnet, 44)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(interval44YAML), 0644); err != nil {
		t.Fatal(err)
	}

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		t.Errorf("unexpected HTTP request: %s", r.URL.Path)
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	overrideEventFileURL(t, server.URL+"/%s/%s")

	event, ok, err := getIntervalRewardsEvent(44, networkMainnet, cacheDir)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	assertBig(t, "executionBlock", event.ExecutionBlock, "24238059")
	if hits.Load() != 0 {
		t.Fatalf("HTTP was called %d times", hits.Load())
	}
}

func TestGetIntervalRewardsEventDownloadAndCache(t *testing.T) {
	for _, network := range []string{networkMainnet, networkTestnet} {
		t.Run(network, func(t *testing.T) {
			var hits atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				want := "/" + network + "/rp-rewards-event-44.yaml"
				if r.URL.Path != want {
					t.Errorf("path = %s, want %s", r.URL.Path, want)
					http.NotFound(w, r)
					return
				}
				if _, err := w.Write([]byte(interval44YAML)); err != nil {
					t.Errorf("write response: %v", err)
				}
			}))
			t.Cleanup(server.Close)
			overrideEventFileURL(t, server.URL+"/%s/%s")

			cacheDir := t.TempDir()
			event, ok, err := getIntervalRewardsEvent(44, network, cacheDir)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if !ok {
				t.Fatal("expected download hit")
			}
			assertBig(t, "executionBlock", event.ExecutionBlock, "24238059")

			cached := rewardsEventLocalPath(cacheDir, network, 44)
			if _, err := os.Stat(cached); err != nil {
				t.Fatalf("expected cached file at %s: %v", cached, err)
			}

			// Second call should be a cache hit.
			_, ok, err = getIntervalRewardsEvent(44, network, cacheDir)
			if err != nil || !ok {
				t.Fatalf("second call: ok=%v err=%v", ok, err)
			}
			if hits.Load() != 1 {
				t.Fatalf("HTTP hits = %d, want 1", hits.Load())
			}
		})
	}
}

func TestGetIntervalRewardsEventNotFoundIsMiss(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	overrideEventFileURL(t, server.URL+"/%s/%s")

	_, ok, err := getIntervalRewardsEvent(44, networkTestnet, t.TempDir())
	if err != nil {
		t.Fatalf("404 should be a miss, not an error: %v", err)
	}
	if ok {
		t.Fatal("expected miss on 404")
	}
}

func TestGetIntervalRewardsEventDownloadFailureIsMiss(t *testing.T) {
	overrideEventFileURL(t, "http://127.0.0.1:1/%s/%s")
	origTimeouts := rewardsEventDownloadTimeouts
	rewardsEventDownloadTimeouts = []time.Duration{20 * time.Millisecond}
	t.Cleanup(func() { rewardsEventDownloadTimeouts = origTimeouts })

	_, ok, err := getIntervalRewardsEvent(44, networkMainnet, t.TempDir())
	if err != nil {
		t.Fatalf("download failure should be a miss, not an error: %v", err)
	}
	if ok {
		t.Fatal("expected miss on download failure")
	}
}

func TestGetIntervalRewardsEventInvalidRemoteYAML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("not: [valid")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	overrideEventFileURL(t, server.URL+"/%s/%s")

	_, ok, err := getIntervalRewardsEvent(44, networkMainnet, t.TempDir())
	if err == nil {
		t.Fatal("expected parse error for invalid remote yaml")
	}
	if ok {
		t.Fatal("invalid yaml must not count as a hit")
	}
}

func overrideEventFileURL(t *testing.T, format string) {
	t.Helper()
	orig := rewardsEventFileURLFormat
	rewardsEventFileURLFormat = format
	t.Cleanup(func() { rewardsEventFileURLFormat = orig })
}

func assertBig(t *testing.T, name string, got *big.Int, want string) {
	t.Helper()
	if got == nil || got.String() != want {
		t.Fatalf("%s = %v, want %s", name, got, want)
	}
}
