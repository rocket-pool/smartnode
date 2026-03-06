package assets

import (
	"bytes"
	_ "embed"
	"io"

	"golang.org/x/crypto/openpgp"
)

//go:embed fornax-signing-key.asc
var fornaxSigningKey []byte
var entityList openpgp.EntityList

func init() {
	var err error
	entityList, err = openpgp.ReadArmoredKeyRing(bytes.NewReader(fornaxSigningKey))
	if err != nil {
		panic(err)
	}
}

func VerifySignedBinary(binary io.Reader, signature io.Reader) (signer *openpgp.Entity, err error) {
	return openpgp.CheckDetachedSignature(entityList, binary, signature)
}
