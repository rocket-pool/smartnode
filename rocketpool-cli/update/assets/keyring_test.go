package assets

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// The fixtures were signed independently with GnuPG 2.4.4 using an RSA-3072
// test key and SHA-512, matching the release signing format. The test key and
// signature were created at 2026-01-01T00:00:00Z without an expiration date.
// Only the public key is retained; no production signing key is used.
func TestVerifySignedBinary(t *testing.T) {
	readFixture := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	payload := readFixture("signed-message")
	signature := readFixture("signed-message.sig")
	testKeys, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(readFixture("signing-key.asc")))
	if err != nil {
		t.Fatal(err)
	}

	// VerifySignedBinary uses the embedded keyring. These subtests must remain
	// sequential while temporarily substituting a test trust root.
	releaseKeys := entityList
	t.Cleanup(func() { entityList = releaseKeys })
	entityList = testKeys

	t.Run("valid GnuPG signature", func(t *testing.T) {
		var downloaded bytes.Buffer
		signer, err := VerifySignedBinary(io.TeeReader(bytes.NewReader(payload), &downloaded), bytes.NewReader(signature))
		if err != nil {
			t.Fatalf("verify signed payload: %v", err)
		}
		if signer == nil || !bytes.Equal(signer.PrimaryKey.Fingerprint, testKeys[0].PrimaryKey.Fingerprint) {
			t.Fatal("verification did not return the trusted signer")
		}
		if !bytes.Equal(downloaded.Bytes(), payload) {
			t.Fatal("verification did not consume the complete binary stream")
		}
	})

	corruptedPayload := bytes.Clone(payload)
	corruptedPayload[0] ^= 1
	corruptedSignature := bytes.Clone(signature)
	corruptedSignature[len(corruptedSignature)-1] ^= 1

	for _, tt := range []struct {
		name      string
		payload   []byte
		signature []byte
	}{
		{"modified binary", corruptedPayload, signature},
		{"truncated binary", payload[:len(payload)-1], signature},
		{"appended binary data", append(bytes.Clone(payload), 0), signature},
		{"empty binary", nil, signature},
		{"corrupted signature", payload, corruptedSignature},
		{"truncated signature", payload, signature[:len(signature)/2]},
		{"empty signature", payload, nil},
		{"malformed signature", payload, []byte("not an OpenPGP signature")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifySignedBinary(bytes.NewReader(tt.payload), bytes.NewReader(tt.signature)); err == nil {
				t.Fatal("verification accepted invalid input")
			}
		})
	}

	t.Run("binary read failure", func(t *testing.T) {
		readErr := errors.New("binary download interrupted")
		reader := io.MultiReader(bytes.NewReader(payload[:5]), signatureErrorReader{readErr})
		if _, err := VerifySignedBinary(reader, bytes.NewReader(signature)); !errors.Is(err, readErr) {
			t.Fatalf("expected binary read error, got %v", err)
		}
	})

	t.Run("signature read failure", func(t *testing.T) {
		readErr := errors.New("signature download interrupted")
		if _, err := VerifySignedBinary(bytes.NewReader(payload), signatureErrorReader{readErr}); !errors.Is(err, readErr) {
			t.Fatalf("expected signature read error, got %v", err)
		}
	})

	t.Run("untrusted signer", func(t *testing.T) {
		entityList = releaseKeys
		t.Cleanup(func() { entityList = testKeys })
		if _, err := VerifySignedBinary(bytes.NewReader(payload), bytes.NewReader(signature)); err == nil {
			t.Fatal("release keyring accepted a signature from an unrelated key")
		}
	})
}

type signatureErrorReader struct {
	err error
}

func (r signatureErrorReader) Read([]byte) (int, error) {
	return 0, r.err
}
