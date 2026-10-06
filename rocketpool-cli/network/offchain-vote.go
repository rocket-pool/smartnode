package network

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// All RocketDash ballot types use one-based choice indices, including ranked
// and approval ballots. Keep ranked ballots in the order submitted by the voter.
func formatOffchainVote(ballot any, choices []string) string {
	label := func(value any) string {
		index, ok := value.(float64)
		if !ok || index < 1 || index > float64(len(choices)) || index != float64(int(index)) {
			return fmt.Sprintf("Unknown (%v)", value)
		}
		return choices[int(index)-1]
	}
	switch ballot := ballot.(type) {
	case float64:
		return label(ballot)
	case []any:
		labels := make([]string, len(ballot))
		for i, choice := range ballot {
			labels[i] = label(choice)
		}
		return strings.Join(labels, ", ")
	case map[string]any:
		keys := make([]string, 0, len(ballot))
		for key := range ballot {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, _ := strconv.Atoi(keys[i])
			b, _ := strconv.Atoi(keys[j])
			if a == b {
				return keys[i] < keys[j]
			}
			return a < b
		})
		labels := make([]string, 0, len(keys))
		for _, key := range keys {
			index, err := strconv.Atoi(key)
			weight, ok := ballot[key].(float64)
			if err != nil || !ok || index < 1 || index > len(choices) {
				labels = append(labels, fmt.Sprintf("Unknown (%s: %v)", key, ballot[key]))
				continue
			}
			labels = append(labels, fmt.Sprintf("%s: %.2f", choices[index-1], weight))
		}
		return strings.Join(labels, ", ")
	default:
		return fmt.Sprintf("Unknown (%v)", ballot)
	}
}
