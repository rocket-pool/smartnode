package client

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/goccy/go-json"
)

// TestValidatorSlicePoolNewReturnsPointer ensures the pool's New function
// returns a *[]string. This is the shape required by staticcheck's SA6002 so
// that sync.Pool.Put does not heap-allocate to box a non-pointer-like slice
// header into an any.
func TestValidatorSlicePoolNewReturnsPointer(t *testing.T) {
	v := validatorSlicePool.New()
	p, ok := v.(*[]string)
	if !ok {
		t.Fatalf("validatorSlicePool.New() returned %T; want *[]string", v)
	}
	if got := cap(*p); got < 1024 {
		t.Fatalf("pooled slice has cap %d; want >= 1024", got)
	}
	if got := len(*p); got != 0 {
		t.Fatalf("pooled slice has len %d; want 0", got)
	}
}

// TestCommitteeUnmarshalPopulatesFields checks that UnmarshalJSON, which
// sources its Validators buffer from validatorSlicePool, correctly decodes
// every field of a Committee.
func TestCommitteeUnmarshalPopulatesFields(t *testing.T) {
	body := []byte(`{"index":"7","slot":"42","validators":["0","1","2","3"]}`)

	var c Committee
	if err := json.Unmarshal(body, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Index != 7 {
		t.Fatalf("Index = %d; want 7", c.Index)
	}
	if c.Slot != 42 {
		t.Fatalf("Slot = %d; want 42", c.Slot)
	}
	if !reflect.DeepEqual(c.Validators, []string{"0", "1", "2", "3"}) {
		t.Fatalf("Validators = %v; want [0 1 2 3]", c.Validators)
	}
}

// TestCommitteesResponseReleaseClearsValidators verifies that Release detaches
// the pooled backing array from the Committee so that callers cannot mutate a
// slice that has been returned to the pool.
func TestCommitteesResponseReleaseClearsValidators(t *testing.T) {
	body := []byte(`{"index":"1","slot":"2","validators":["a","b","c"]}`)

	var c Committee
	if err := json.Unmarshal(body, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	resp := &CommitteesResponse{Data: []Committee{c}}
	resp.Release()

	if resp.Data[0].Validators != nil {
		t.Fatalf("Validators = %v; want nil after Release", resp.Data[0].Validators)
	}
}

// TestValidatorSlicePoolRoundTripReusesBackingArray confirms that Release puts
// the slice back into the pool and the subsequent UnmarshalJSON re-uses the
// same backing array. This verifies the whole point of the pool: avoiding
// re-allocation of large validator slices.
//
// sync.Pool does not formally guarantee that a Put item is returned by the
// next Get, but within a single goroutine pinned to its P without any
// intervening GC, Go's per-P cache will service the Get from the most recent
// Put. We lock the OS thread and avoid runtime.GC calls to make this
// deterministic for the test.
func TestValidatorSlicePoolRoundTripReusesBackingArray(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	body := []byte(`{"index":"0","slot":"0","validators":["x","y","z"]}`)

	var first Committee
	if err := json.Unmarshal(body, &first); err != nil {
		t.Fatalf("first unmarshal: %v", err)
	}
	firstArray := unsafe.SliceData(first.Validators)

	resp := &CommitteesResponse{Data: []Committee{first}}
	resp.Release()

	var second Committee
	if err := json.Unmarshal(body, &second); err != nil {
		t.Fatalf("second unmarshal: %v", err)
	}
	secondArray := unsafe.SliceData(second.Validators)

	if firstArray != secondArray {
		t.Fatalf("backing array was not reused: first=%p second=%p", firstArray, secondArray)
	}
}
