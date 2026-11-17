package node

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

func TestMegapoolValidatorSetEntry(t *testing.T) {
	if got := megapoolValidatorSetKey(4); got != common.HexToHash("0x940489ed09426b7e471541a102b1e86fb32c606af546fb7053e52a499aa4d580") {
		t.Fatalf("validator set key = %s", got.Hex())
	}
	address := common.HexToAddress("0x0DFB735d953453198D0Cd79546464455b40D3645")
	encoded := new(big.Int).Lsh(new(big.Int).SetBytes(address.Bytes()), 96)
	encoded.Or(encoded, big.NewInt(7))
	got, err := decodeMegapoolValidatorSetEntry(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got != address {
		t.Fatalf("decoded megapool = %s", got.Hex())
	}
	if _, err := decodeMegapoolValidatorSetEntry(nil); err == nil {
		t.Fatal("nil entry decoded")
	}
	if _, err := decodeMegapoolValidatorSetEntry(big.NewInt(0)); err == nil {
		t.Fatal("empty entry decoded")
	}
}

func TestMegapoolDelegateNeedsUpgrade(t *testing.T) {
	if megapoolDelegateNeedsUpgrade(false, false) {
		t.Fatal("a current delegate does not need an upgrade")
	}
	if megapoolDelegateNeedsUpgrade(true, false) {
		t.Fatal("use-latest with a current delegate does not need an upgrade")
	}
	if megapoolDelegateNeedsUpgrade(true, true) {
		t.Fatal("use-latest does not pin an expired delegate")
	}
	if !megapoolDelegateNeedsUpgrade(false, true) {
		t.Fatal("an expired pinned delegate must be upgraded before getValidatorInfo")
	}
}

func TestClassifyPrestake(t *testing.T) {
	info := megapool.ValidatorPrestakeInfo{
		ValidatorInfo: megapool.ValidatorInfo{InPrestake: true},
		Signature:     []byte{1},
	}
	if got := classifyPrestake(info, true); got != prestakeValid {
		t.Fatalf("valid signature classified as %d", got)
	}
	if got := classifyPrestake(info, false); got != prestakeInvalidSignature {
		t.Fatalf("invalid signature classified as %d", got)
	}
	info.Signature = nil
	if got := classifyPrestake(info, false); got != prestakeSkipped {
		t.Fatalf("missing signature classified as %d", got)
	}
	info.Signature = []byte{1}
	info.ValidatorInfo.InPrestake = false
	if got := classifyPrestake(info, false); got != prestakeSkipped {
		t.Fatalf("validator not in prestake classified as %d", got)
	}
}

func TestPrestakePageAdvances(t *testing.T) {
	if prestakePageAdvances(big.NewInt(0), big.NewInt(0)) {
		t.Fatal("cursor did not move")
	}
	if !prestakePageAdvances(big.NewInt(0), big.NewInt(prestakeIndexPageSize)) {
		t.Fatal("cursor advanced")
	}
	if prestakePageAdvances(big.NewInt(10), nil) {
		t.Fatal("nil cursor advanced")
	}
}

func TestPrestakeSignatureValid(t *testing.T) {
	secret, err := bls.RandKey()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := secret.PublicKey().Marshal()
	megapoolAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	credentials := services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)
	domain, err := signing.ComputeDomain(eth2types.DomainDeposit, []byte{0, 0, 0, 0}, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		t.Fatal(err)
	}

	signature, err := signDeposit(secret, pubkey, credentials, prestakeAmountGwei, domain)
	if err != nil {
		t.Fatal(err)
	}
	if !prestakeSignatureValid(pubkey, signature, credentials, domain) {
		t.Fatal("expected the 1 ETH megapool deposit signature to verify")
	}

	otherCredentials := services.CalculateMegapoolWithdrawalCredentials(common.HexToAddress("0x2222222222222222222222222222222222222222"))
	if prestakeSignatureValid(pubkey, signature, otherCredentials, domain) {
		t.Fatal("expected a signature over different withdrawal credentials to be invalid")
	}

	wrongAmount, err := signDeposit(secret, pubkey, credentials, 32_000_000_000, domain)
	if err != nil {
		t.Fatal(err)
	}
	if prestakeSignatureValid(pubkey, wrongAmount, credentials, domain) {
		t.Fatal("expected a signature over 32 ETH to be invalid")
	}
	if prestakeSignatureValid(pubkey, nil, credentials, domain) {
		t.Fatal("expected an empty signature to be invalid")
	}
}

func TestWrongCredentialPendingDeposits(t *testing.T) {
	secret, err := bls.RandKey()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := secret.PublicKey().Marshal()
	megapoolAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	megapoolCredentials := services.CalculateMegapoolWithdrawalCredentials(megapoolAddress)
	wrongCredentials := services.CalculateMegapoolWithdrawalCredentials(common.HexToAddress("0x2222222222222222222222222222222222222222"))
	domain, err := signing.ComputeDomain(eth2types.DomainDeposit, []byte{0, 0, 0, 0}, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		t.Fatal(err)
	}
	const amount = uint64(2_000_000_000)
	signature, err := signDeposit(secret, pubkey, wrongCredentials, amount, domain)
	if err != nil {
		t.Fatal(err)
	}
	deposit := &generic.PendingDeposit{
		Pubkey:                pubkey,
		WithdrawalCredentials: append([]byte(nil), wrongCredentials.Bytes()...),
		Amount:                amount,
		Signature:             signature,
		Slot:                  3,
	}
	info := megapool.ValidatorPrestakeInfo{
		Pubkey:          pubkey,
		MegapoolAddress: megapoolAddress,
		ValidatorId:     4,
		ValidatorInfo:   megapool.ValidatorInfo{InPrestake: true},
	}

	matches := wrongCredentialPendingDeposits([]*generic.PendingDeposit{deposit}, []megapool.ValidatorPrestakeInfo{info}, domain)
	if len(matches) != 1 || matches[0].depositIndex != 0 || matches[0].info.ValidatorId != 4 {
		t.Fatalf("expected the wrong-credential deposit to dissolve validator 4, got %+v", matches)
	}

	goodSignature, err := signDeposit(secret, pubkey, megapoolCredentials, amount, domain)
	if err != nil {
		t.Fatal(err)
	}
	correct := *deposit
	correct.WithdrawalCredentials = append([]byte(nil), megapoolCredentials.Bytes()...)
	correct.Signature = goodSignature
	if got := wrongCredentialPendingDeposits([]*generic.PendingDeposit{&correct}, []megapool.ValidatorPrestakeInfo{info}, domain); len(got) != 0 {
		t.Fatalf("megapool credentials were treated as invalid: %+v", got)
	}

	wrongAmount := *deposit
	wrongAmount.Amount = amount + 1
	if got := wrongCredentialPendingDeposits([]*generic.PendingDeposit{&wrongAmount}, []megapool.ValidatorPrestakeInfo{info}, domain); len(got) != 0 {
		t.Fatalf("an invalid signature was selected: %+v", got)
	}

	otherPubkey := *deposit
	otherPubkey.Pubkey = bytes.Repeat([]byte{9}, 48)
	if got := wrongCredentialPendingDeposits([]*generic.PendingDeposit{&otherPubkey}, []megapool.ValidatorPrestakeInfo{info}, domain); len(got) != 0 {
		t.Fatalf("a different pubkey was selected: %+v", got)
	}

	notPrestake := info
	notPrestake.ValidatorInfo.InPrestake = false
	if got := wrongCredentialPendingDeposits([]*generic.PendingDeposit{deposit}, []megapool.ValidatorPrestakeInfo{notPrestake}, domain); len(got) != 0 {
		t.Fatalf("a validator that is not prestaked was selected: %+v", got)
	}

	second := *deposit
	matches = wrongCredentialPendingDeposits([]*generic.PendingDeposit{deposit, &second}, []megapool.ValidatorPrestakeInfo{info}, domain)
	if len(matches) != 1 || matches[0].depositIndex != 0 {
		t.Fatalf("expected one dissolve for the first matching deposit, got %+v", matches)
	}
}

func signDeposit(secret bls.SecretKey, pubkey []byte, credentials common.Hash, amount uint64, domain []byte) ([]byte, error) {
	message := &ethpb.DepositMessage{
		PublicKey:             pubkey,
		WithdrawalCredentials: credentials.Bytes(),
		Amount:                amount,
	}
	root, err := message.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	signingData := &ethpb.SigningData{
		ObjectRoot: root[:],
		Domain:     domain,
	}
	signingRoot, err := signingData.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	return secret.Sign(signingRoot[:]).Marshal(), nil
}
