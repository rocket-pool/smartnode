package types

import (
    "fmt"

    "encoding/hex"
    "encoding/json"
)


// Validator pubkey
const ValidatorPubkeyLength = 48 // bytes
type ValidatorPubkey [ValidatorPubkeyLength]byte


// Bytes conversion
func (v ValidatorPubkey) Bytes() []byte {
    return v[:]
}
func BytesToValidatorPubkey(value []byte) ValidatorPubkey {
    var pubkey ValidatorPubkey
    copy(pubkey[:], value)
    return pubkey
}


// String conversion
func (v ValidatorPubkey) Hex() string {
    return hex.EncodeToString(v.Bytes())
}
func (v ValidatorPubkey) String() string {
    return v.Hex()
}
func HexToValidatorPubkey(value string) (ValidatorPubkey, error) {
    pubkey := make([]byte, ValidatorPubkeyLength)
    if len(value) != hex.EncodedLen(ValidatorPubkeyLength) {
        return ValidatorPubkey{}, fmt.Errorf("Invalid validator public key hex string %s: invalid length %d", value, len(value))
    }
    if _, err := hex.Decode(pubkey, []byte(value)); err != nil {
        return ValidatorPubkey{}, err
    }
    return BytesToValidatorPubkey(pubkey), nil
}


// JSON encoding
func (v ValidatorPubkey) MarshalJSON() ([]byte, error) {
    return json.Marshal(v.Hex())
}
func (v *ValidatorPubkey) UnmarshalJSON(data []byte) error {
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil { return err }
    pubkey, err := HexToValidatorPubkey(dataStr)
    if err == nil { *v = pubkey }
    return err
}


// Validator signature
const ValidatorSignatureLength = 96 // bytes
type ValidatorSignature [ValidatorSignatureLength]byte


// Bytes conversion
func (v ValidatorSignature) Bytes() []byte {
    return v[:]
}
func BytesToValidatorSignature(value []byte) ValidatorSignature {
    var signature ValidatorSignature
    copy(signature[:], value)
    return signature
}


// String conversion
func (v ValidatorSignature) Hex() string {
    return hex.EncodeToString(v.Bytes())
}
func (v ValidatorSignature) String() string {
    return v.Hex()
}
func HexToValidatorSignature(value string) (ValidatorSignature, error) {
    signature := make([]byte, ValidatorSignatureLength)
    if len(value) != hex.EncodedLen(ValidatorSignatureLength) {
        return ValidatorSignature{}, fmt.Errorf("Invalid validator signature hex string %s: invalid length %d", value, len(value))
    }
    if _, err := hex.Decode(signature, []byte(value)); err != nil {
        return ValidatorSignature{}, err
    }
    return BytesToValidatorSignature(signature), nil
}


// JSON encoding
func (v ValidatorSignature) MarshalJSON() ([]byte, error) {
    return json.Marshal(v.Hex())
}
func (v *ValidatorSignature) UnmarshalJSON(data []byte) error {
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil { return err }
    signature, err := HexToValidatorSignature(dataStr)
    if err == nil { *v = signature }
    return err
}

