package ssz_types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

var networkMap = map[string]Network{
	"mainnet": 1,
	"holesky": 17000,
}

// internal use only
type sszfile_v1_alias SSZFile_v1

// This custom unmarshaler avoids creating a landmine where the user
// may forget to call NewSSZFile_v1 before unmarshaling into the result,
// which would cause the Magic header to be unset.
func (f *SSZFile_v1) UnmarshalJSON(data []byte) error {
	// Disposable type without a custom unmarshal
	var alias sszfile_v1_alias
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	*f = SSZFile_v1(alias)

	// After unmarshaling, set the magic header
	f.Magic = Magic

	// Verify legitimacy of the file
	return f.Verify()
}

// When writing JSON, we need to compute the merkle tree to populate the proofs
func (f *SSZFile_v1) MarshalJSON() ([]byte, error) {
	if err := f.Verify(); err != nil {
		return nil, fmt.Errorf("error verifying ssz while serializing json: %w", err)
	}
	proofs, err := f.Proofs()
	if err != nil {
		return nil, fmt.Errorf("error getting proofs: %w", err)
	}

	for _, nr := range f.NodeRewards {
		proof, ok := proofs[nr.Address]
		if !ok {
			return nil, fmt.Errorf("error getting proof for node %s", nr.Address)
		}
		nr.MerkleProof = proof
	}

	var alias sszfile_v1_alias
	alias = sszfile_v1_alias(*f)
	return json.Marshal(&alias)
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	s = strings.TrimPrefix(s, "0x")
	out, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	if len(out) != 32 {
		return fmt.Errorf("merkle root %s wrong size- must be 32 bytes", s)
	}

	copy((*[32]byte)(h)[:], out)
	return nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}

func NetworkFromString(s string) (Network, bool) {
	n, ok := networkMap[s]
	return n, ok
}

func (n *Network) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	id, ok := NetworkFromString(s)
	if ok {
		*n = Network(id)
		return nil
	}

	// If the network string doesn't match known values, try to treat it as an integer
	u, err := strconv.ParseUint(s, 10, 64)
	if err == nil {
		*n = Network(u)
		return nil
	}

	// If the network string isn't an integer, use UINT64_MAX
	*n = Network(math.MaxUint64)
	return nil
}

func (n Network) MarshalJSON() ([]byte, error) {
	id := n
	for k, v := range networkMap {
		if v == id {
			return json.Marshal(k)
		}
	}

	// If the network id isn't in the map, serialize it as a string
	return json.Marshal(strconv.FormatUint(uint64(id), 10))
}

func (n *NetworkRewards) UnmarshalJSON(data []byte) error {
	// Network Rewards is a slice, but represented as a map in the json.
	var m map[string]json.RawMessage
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	*n = make(NetworkRewards, 0, len(m))
	for k, v := range m {
		networkId, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		networkReward := new(NetworkReward)
		networkReward.Network = networkId

		err = json.Unmarshal(v, networkReward)
		if err != nil {
			return err
		}
		*n = append(*n, networkReward)
	}

	sort.Sort(*n)
	return nil
}

func (n NetworkRewards) MarshalJSON() ([]byte, error) {
	// Network Rewards is a slice, but represented as a map in the json.
	m := make(map[string]*NetworkReward, len(n))
	// Make sure we sort, first
	sort.Sort(n)
	for _, nr := range n {
		m[strconv.FormatUint(nr.Network, 10)] = nr
	}

	// Serialize the map
	return json.Marshal(m)
}

func (n *NodeRewards) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	*n = make(NodeRewards, 0, len(m))
	for k, v := range m {
		s := strings.TrimPrefix(k, "0x")
		addr, err := hex.DecodeString(s)
		if err != nil {
			return err
		}

		if len(addr) != 20 {
			return fmt.Errorf("address %s wrong size- must be 20 bytes", s)
		}

		nodeReward := new(NodeReward)
		copy(nodeReward.Address[:], addr)
		err = json.Unmarshal(v, nodeReward)
		if err != nil {
			return err
		}
		*n = append(*n, nodeReward)
	}

	sort.Sort(*n)
	return nil
}

func (n NodeRewards) MarshalJSON() ([]byte, error) {
	// Node Rewards is a slice, but represented as a map in the json.
	m := make(map[string]*NodeReward, len(n))
	// Make sure we sort, first
	sort.Sort(n)
	for _, nr := range n {
		m[nr.Address.String()] = nr
	}

	// Serialize the map
	return json.Marshal(m)
}
