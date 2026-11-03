# Smartnode Safe Recovery Planner

This package implements the **two-phase (Plan-Before-Write) recovery architecture** and per-key diagnosis for Rocket Pool Smartnode validator keys, addressing issues **#407**, **#440**, and **#572**.

---

## The Problem

Previously, Smartnode's key recovery routines (`wallet recover`, `wallet rebuild`, and `wallet test-recovery`) immediately persisted keys to disk as soon as each individual key was discovered. In mixed-key deployments (a combination of mnemonic-derived keys, imported custom EIP-2335 keystores, legacy minipools, and megapool validators), if recovery succeeded for several keys but subsequently failed on a later key (e.g., missing custom keystore, corrupt file, invalid password, or derivation limit exceeded), the node was left in an inconsistent, partially-recovered state.

Additionally, inactive/exited validators on the Beacon chain were filtered into an active list, but the target discovery map was populated from the original unfiltered list, causing exited validators to trigger false recovery failures (Issue #407).

---

## Architectural Principles

### 1. Two-Phase Execution (Plan-Before-Write)

Recovery is strictly separated into two independent phases:

1. **Plan Phase (`PlanValidatorRecovery`)**:
   - Pure, read-only discovery and diagnosis.
   - Evaluates all expected minipool and megapool public keys against Beacon chain validator states, installed client keystores, imported custom keys (`custom_keys/`), and mnemonic derivation limits.
   - Assigns a deterministic `DiscoveryOutcome` to every single public key.
   - **Zero file mutations, zero keystore modifications.**

2. **Commit Phase (`CommitPlan`)**:
   - Evaluates the safety invariants of the plan before touching disk.
   - By default, **any** unresolved or invalid in-scope validator key causes the commit phase to abort immediately before any keystore is touched (`ErrUnresolvedKeysRefusingCommit`).
   - Only when all in-scope keys are resolved (or when the operator explicitly passes `--allow-partial-recover`), the validated keys are committed to disk via `KeyWriter`.

### 2. Deterministic Per-Key Outcomes

Each public key is assigned a deterministic diagnosis:
- **`mnemonic_valid`**: Successfully derived from the node mnemonic and verified against the public key.
- **`custom_valid`**: Loaded from an imported EIP-2335 keystore and decrypted with the provided password.
- **`installed_valid`**: Key material already exists and is valid in the node's validator client keystores.
- **`unresolved`**: Not found in custom keys or within the derivation search window (`bucketLimit`).
- **`invalid_material`**: Corrupt keystore, incorrect password, or pubkey/private key mismatch.
- **`excluded_inactive`**: Exited or inactive on the Beacon chain; safely excluded from recovery.

Commit status is tracked explicitly:
- `written`: Successfully written to keystores.
- `skipped_already_installed`: Key was already installed; avoided redundant overwrite.
- `skipped_excluded`: Inactive validator; no keys written.
- `skipped_partial`: Unresolved key omitted during an operator-approved partial recovery.
- `not_attempted`: Write phase was blocked because plan was invalid or unapproved.

### 3. Partial Recovery Behavior (`--allow-partial-recover`)

When an operator cannot locate a lost custom key or wants to bring online only the validators that could be found, passing `--allow-partial-recover` allows Smartnode to commit all resolved keys while recording which keys were skipped (`skipped_partial`). This prevents operators from being stuck when one obsolete or migrated key cannot be derived.

### 4. Exclusion of Cross-Client Rollback

Smartnode interfaces with diverse external Validator Clients (Lighthouse, Nimbus, Prysm, Teku, Lodestar), each with its own local keystore formats and disk layouts. This architecture provides a **pre-mutation safety guarantee** (ensuring all keys are validated in memory before any writes begin). It does not attempt transactional cross-client atomic rollback in the event of hardware or OS-level disk failures mid-write. Operators can safely re-run recovery idempotently at any time.
