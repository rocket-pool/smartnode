package performance

import (
	"math/big"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestChallengeRequestParsing(t *testing.T) {
	huge := new(big.Int).Lsh(big.NewInt(1), 255).String()
	for _, test := range []struct {
		name   string
		mini   bool
		values url.Values
		valid  bool
	}{
		{
			name:   "megapool batch",
			mini:   false,
			values: url.Values{"validatorIds": {"0,7,4294967295"}, "startEpoch": {"123"}, "participation": {huge + ",0"}},
			valid:  true,
		},
		{
			name:   "legacy scalar",
			mini:   false,
			values: url.Values{"validatorId": {"0"}, "startEpoch": {"123"}, "participation": {"1"}},
			valid:  true,
		},
		{
			name:   "overflow id",
			mini:   false,
			values: url.Values{"validatorIds": {"4294967296"}, "startEpoch": {"123"}, "participation": {"1"}},
			valid:  false,
		},
		{
			name:   "negative word",
			mini:   false,
			values: url.Values{"validatorIds": {"0"}, "startEpoch": {"123"}, "participation": {"-1"}},
			valid:  false,
		},
		{
			name:   "overflow word",
			mini:   false,
			values: url.Values{"validatorIds": {"0"}, "startEpoch": {"123"}, "participation": {new(big.Int).Lsh(big.NewInt(1), 256).String()}},
			valid:  false,
		},
		{
			name:   "duplicate id",
			mini:   false,
			values: url.Values{"validatorIds": {"0,0"}, "startEpoch": {"123"}, "participation": {"1"}},
			valid:  false,
		},
		{
			name:   "missing epoch",
			mini:   false,
			values: url.Values{"validatorIds": {"0"}, "participation": {"1"}},
			valid:  false,
		},
		{
			name:   "minipool list",
			mini:   true,
			values: url.Values{"minipoolAddresses": {"0x0000000000000000000000000000000000000001,0x0000000000000000000000000000000000000002"}, "startEpoch": {"123"}, "participation": {"1"}},
			valid:  true,
		},
		{
			name:   "zero minipool",
			mini:   true,
			values: url.Values{"minipoolAddresses": {"0x0000000000000000000000000000000000000000"}, "startEpoch": {"123"}, "participation": {"1"}},
			valid:  false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, method := range []string{"GET", "POST"} {
				request := httptest.NewRequest(method, "/?"+test.values.Encode(), nil)
				if method == "POST" {
					request = httptest.NewRequest(method, "/", strings.NewReader(test.values.Encode()))
					request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				}
				got, err := parseRequest(request, test.mini)
				if (err == nil) != test.valid {
					t.Fatalf("%s: valid=%t, err=%v", method, test.valid, err)
				}
				if test.name == "megapool batch" && (len(got.ValidatorIds) != 3 || got.ValidatorIds[2] != ^uint32(0) || got.Participation[0].String() != huge) {
					t.Fatalf("payload truncated: %+v", got)
				}
			}
		})
	}
}
