// Copyright 2020 Prysmatic Labs
// This file is part of the Prysm ethereum 2.0 client.
//
// Modified by Rocket Pool 2020
// (from Prysm v1.0.0-alpha: /beacon-chain/core/helpers/signing_root.go)

package bls

import (
	"github.com/prysmaticlabs/go-ssz"
)

// ForkVersionByteLength length of fork version byte array.
const ForkVersionByteLength = 4

// DomainByteLength length of domain byte array.
const DomainByteLength = 4

// Default genesisValidatorsRoot
var ZeroHash = [32]byte{}

// Fork data
type ForkData struct {
	CurrentVersion []byte
	GenesisValidatorsRoot []byte
}

// ComputeDomain returns the domain version for BLS private key to sign and verify with a zeroed 4-byte
// array as the fork version.
//
// def compute_domain(domain_type: DomainType, fork_version: Version=None, genesis_validators_root: Root=None) -> Domain:
//    """
//    Return the domain for the ``domain_type`` and ``fork_version``.
//    """
//    if fork_version is None:
//        fork_version = GENESIS_FORK_VERSION
//    if genesis_validators_root is None:
//        genesis_validators_root = Root()  # all bytes zero by default
//    fork_data_root = compute_fork_data_root(fork_version, genesis_validators_root)
//    return Domain(domain_type + fork_data_root[:28])
func ComputeDomain(domainType [DomainByteLength]byte, forkVersion []byte, genesisValidatorsRoot []byte) ([]byte, error) {
	if genesisValidatorsRoot == nil {
		genesisValidatorsRoot = ZeroHash[:]
	}
	forkBytes := [ForkVersionByteLength]byte{}
	copy(forkBytes[:], forkVersion)

	forkDataRoot, err := computeForkDataRoot(forkBytes[:], genesisValidatorsRoot)
	if err != nil {
		return nil, err
	}

	return domain(domainType, forkDataRoot[:]), nil
}

// This returns the bls domain given by the domain type and fork data root.
func domain(domainType [DomainByteLength]byte, forkDataRoot []byte) []byte {
	b := []byte{}
	b = append(b, domainType[:4]...)
	b = append(b, forkDataRoot[:28]...)
	return b
}

// this returns the 32byte fork data root for the ``current_version`` and ``genesis_validators_root``.
// This is used primarily in signature domains to avoid collisions across forks/chains.
//
// Spec pseudocode definition:
//	def compute_fork_data_root(current_version: Version, genesis_validators_root: Root) -> Root:
//    """
//    Return the 32-byte fork data root for the ``current_version`` and ``genesis_validators_root``.
//    This is used primarily in signature domains to avoid collisions across forks/chains.
//    """
//    return hash_tree_root(ForkData(
//        current_version=current_version,
//        genesis_validators_root=genesis_validators_root,
//    ))
func computeForkDataRoot(version []byte, root []byte) ([32]byte, error) {
	r, err := ssz.HashTreeRoot(ForkData{
		CurrentVersion:        version,
		GenesisValidatorsRoot: root,
	})
	if err != nil {
		return [32]byte{}, err
	}
	return r, nil
}
