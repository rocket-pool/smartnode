package client

import (
	"fmt"
	"sync"

	"github.com/goccy/go-json"
)

type Committee struct {
	Index      uinteger `json:"index"`
	Slot       uinteger `json:"slot"`
	Validators []string `json:"validators"`
}

// Custom deserialization logic for Committee allows us to pool the validator
// slices for reuse. They're quite large, so this cuts down on allocations
// substantially.
var validatorSlicePool sync.Pool = sync.Pool{
	New: func() any {
		out := make([]string, 0, 1024)
		return &out
	},
}

func (c *Committee) UnmarshalJSON(body []byte) error {
	var committee map[string]*json.RawMessage

	pooledSlicePtr := validatorSlicePool.Get().(*[]string)
	c.Validators = *pooledSlicePtr

	// Partially parse the json
	if err := json.Unmarshal(body, &committee); err != nil {
		return fmt.Errorf("error unmarshalling committee json: %w\n", err)
	}

	// Parse each field
	if err := json.Unmarshal(*committee["index"], &c.Index); err != nil {
		return err
	}
	if err := json.Unmarshal(*committee["slot"], &c.Slot); err != nil {
		return err
	}
	// Since c.Validators was preallocated, this will re-use a buffer if one was available.
	if err := json.Unmarshal(*committee["validators"], &c.Validators); err != nil {
		return err
	}

	return nil
}

func (c *CommitteesResponse) Count() int {
	return len(c.Data)
}

func (c *CommitteesResponse) Index(idx int) uint64 {
	return uint64(c.Data[idx].Index)
}

func (c *CommitteesResponse) Slot(idx int) uint64 {
	return uint64(c.Data[idx].Slot)
}

func (c *CommitteesResponse) Validators(idx int) []string {
	return c.Data[idx].Validators
}

func (c *CommitteesResponse) Release() {
	for _, committee := range c.Data {
		// Reset the slice length to 0 (capacity stays the same)
		committee.Validators = committee.Validators[:0]
		// Return the slice for reuse
		validatorSlicePool.Put(&committee.Validators)
	}
}
