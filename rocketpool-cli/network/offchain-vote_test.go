package network

import (
	"encoding/json"
	"testing"
)

func TestFormatOffchainVote(t *testing.T) {
	for _, tc := range []struct {
		ballot string
		want   string
	}{
		{`1`, "For"},
		{`[3,1,2]`, "Abstain, For, Against"},
		{`[1,3]`, "For, Abstain"},
		{`{"3":1,"1":2}`, "For: 2.00, Abstain: 1.00"},
		{`0`, "Unknown (0)"},
		{`1.5`, "Unknown (1.5)"},
		{`[4,"bad"]`, "Unknown (4), Unknown (bad)"},
		{`{"0":2,"4":1}`, "Unknown (0: 2), Unknown (4: 1)"},
		{`{"bad":"weight"}`, "Unknown (bad: weight)"},
	} {
		t.Run(tc.ballot, func(t *testing.T) {
			var ballot any
			if err := json.Unmarshal([]byte(tc.ballot), &ballot); err != nil {
				t.Fatal(err)
			}
			if got := formatOffchainVote(ballot, []string{"For", "Against", "Abstain"}); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
