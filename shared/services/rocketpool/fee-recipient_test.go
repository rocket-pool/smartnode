package rocketpool

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadFeeRecipientResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		status      string
		body        string
		wantAddress string
		wantError   string
	}{
		{
			name:        "success",
			statusCode:  http.StatusOK,
			status:      "200 OK",
			body:        `{"data":{"ethaddress":"0x1234"}}`,
			wantAddress: "0x1234",
		},
		{
			name:       "unauthorized response",
			statusCode: http.StatusUnauthorized,
			status:     "401 Unauthorized",
			body:       `{"message":"invalid token"}`,
			wantError:  "401 Unauthorized",
		},
		{
			name:       "malformed success response",
			statusCode: http.StatusOK,
			status:     "200 OK",
			body:       `{`,
			wantError:  "decoding fee recipient response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Status:     tt.status,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			address, err := readFeeRecipientResponse(resp)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantError)
				}
				if address != "" {
					t.Fatalf("address = %q, want empty", address)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if address != tt.wantAddress {
				t.Fatalf("address = %q, want %q", address, tt.wantAddress)
			}
		})
	}
}
