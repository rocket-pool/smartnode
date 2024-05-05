# Changelog

## [v1.13.0](https://github.com/rocket-pool/smartnode/tree/v1.13.0) (2024-05-02)

### Changed

- Geth updated to v1.14.0
- Nethermind updated to v1.26.0
- Teku updated to v24.4.0
- Nimbus updated to v24.4.0
- Node exporter updated to v1.8.0
- Prometheus updated to v2.51.2
- Grafana updated to v9.5.18

### Added

- Added Geth Archive mode option on the TUI
- Added an evmTimeout TUI parameter for Geth

### Fixed

- Fixed Grafana EC peers query - Dashboard v1.3.1
- Fixed Reth RPC cors parameter
- Fixed check for hosts that registered pre-Houston
- Improved the pdao voting-power command
- Improved Snapshot/Onchain messages on node status.
- Fixed starting the SN with alerting enabled and metrics disabled
- Fixed RPL withdraw regression from v1.13.0

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.13.0...v1.13.1)

## [v1.13.0](https://github.com/rocket-pool/smartnode/tree/v1.13.0) (2024-04-23)

### Changed

- Update Geth version
- Update Besu version
- Update Lodestar version
- Update Reth version
- Updated Go to v1.21 (https://github.com/rocket-pool/smartnode/pull/476)

### Added

- Implementation of an On-Chain pDAO. RPIP-33
- Stake ETH on behalf of node. RPIP-32
- RPL Withdrawal Address. RPIP-31
- Time-based Balance and RPL Price Submissions. RPIP-35
- Added a linter CI action to GitHub (https://github.com/rocket-pool/smartnode/pull/490)
- Added a Changelog (https://github.com/rocket-pool/smartnode/pull/504)

### Fixed

- v8-rolling records to have the same behavior as v8 rewards (https://github.com/rocket-pool/smartnode/pull/474)
- Allow scrubbed/dissolved minipools with refunds to be closed (https://github.com/rocket-pool/smartnode/pull/487)
- Fix invalid Smart Node Update available (https://github.com/rocket-pool/smartnode-install/pull/126)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.12.1...v1.13.0)

## [v1.12.1](https://github.com/rocket-pool/smartnode/tree/v1.12.1) (2024-04-02)

### Fixed

- Besu version corrected

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.12.0...v1.12.1)

## [v1.12.0](https://github.com/rocket-pool/smartnode/tree/v1.12.0) (2024-03-30)

### Changed

- Update nethermind version
- Update Reth version
- Update Lighthouse version
- Update Teku version
- Update Prysm version
- Update Nimbus version
- Update Besu version

### Added

- Smartnode notification functionality for bounty BA022308 (https://github.com/rocket-pool/smartnode/pull/449)

### Fixed

- When bc status is not syncing fix progress to 1.0 (https://github.com/rocket-pool/smartnode/pull/469)
- rocketpool_node approximate rpl reward panel for RPIP-30 changes (https://github.com/rocket-pool/smartnode/pull/457)
- Support for using alerting in native mode (https://github.com/rocket-pool/smartnode/pull/465)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.9...v1.12.0)

## [v1.11.9](https://github.com/rocket-pool/smartnode/tree/v1.11.9) (2024-03-11)

### Changed

- Update Reth version
- Update Teku version
- Update Grafana (https://github.com/rocket-pool/smartnode/pull/456)
- Update Besu version
- Update Prysm version
- Update Lighthouse version
- Update Lodestar version

### Added

- Add MEV-Boost version in `rocketpool service version` (https://github.com/rocket-pool/smartnode/pull/455)
- Add --all option for `rocketpool service prune` and `rocketpool service reset` per feedback in :(https://github.com/rocket-pool/smartnode/issues/323)
- Rocketpool service reset for bounty BA0902402 (https://github.com/rocket-pool/smartnode/pull/452)
- Pruning Besu if not archive node
- Check node has more than maxRPLStake on withdrawals

### Fixed

- rocketpool node stake-rpl reads incorrect selection (https://github.com/rocket-pool/smartnode/issues/450)
- Generate state manager when needed (https://github.com/rocket-pool/smartnode/pull/453)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.8...v1.11.9)

## [v1.11.8](https://github.com/rocket-pool/smartnode/tree/v1.11.8) (2024-02-27)

### Changed

- Update Lodestar version
- Update Geth version
- Update Reth version
- Update Besu version
- Update Teku version
- Update Nethermind version
- Update Nimbus version

### Added

- Reth client
- Max stake fraction from contract
- Filter out duplicate pubkeys (https://github.com/rocket-pool/smartnode/pull/443)
- `MaxCollateralFraction` hardcoded to 1.5 eth (https://github.com/rocket-pool/smartnode/pull/445)
- Timeouts to rewards tree downloads (https://github.com/rocket-pool/smartnode/pull/446)
- Besu archive mdoe

### Fixed

- "-c" shorthand for subcommands (https://github.com/rocket-pool/smartnode/pull/441)
- Voluntary exit signatures (https://github.com/rocket-pool/smartnode/pull/447)
- `rewardsFileVersion` type encapsulates version range checks. Errors more verbose (https://github.com/rocket-pool/smartnode/pull/448)

### Removed

- Max rpl stake options

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.7...v1.11.8)

## [v1.11.7](https://github.com/rocket-pool/smartnode/tree/v1.11.7) (2024-02-02)

### Changed

- Update Teku version
- Update Lighthouse version
- Update Besu version
- Update Nethermind version
- Update Geth version
- Update Nimbus version
- Update Lodestar version
- Update Prysm version

### Removed

- Removed Goerli
- Removed Prater (https://github.com/rocket-pool/smartnode/pull/437)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.6...v1.11.7)

## [v1.11.6](https://github.com/rocket-pool/smartnode/tree/v1.11.6) (2024-01-22)

### Changed

- Update Nethermind version
- Update Lodestar version
- Update Nethermind version

### Added

- Test to cement CID calculation (https://github.com/rocket-pool/smartnode/pull/433)
- Not omit totalNodeWeight if not set (https://github.com/rocket-pool/smartnode/pull/433)

### Fixed

- Go version of the Nethermind prune starter, replacing the dll (https://github.com/rocket-pool/smartnode/pull/434)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.5...v1.11.6)

## [v1.11.5](https://github.com/rocket-pool/smartnode/tree/v1.11.5) (2024-01-13)

### Changed

- Update Geth version
- Update Nethermind version
- Update Prysm version
- Update Besu version
- Prysm max peers increased
- Holesky to v8 on interval 93

### Added

- Total node weight printed in treegen logs (https://github.com/rocket-pool/smartnode/pull/431)
- Local file abstraction for treegen (https://github.com/rocket-pool/smartnode/pull/432)
- Arbitrum price messenger v2

### Fixed

- RR CID calculation
- Compression bugs with CID calculation (https://github.com/rocket-pool/smartnode/pull/432)
- CID filename usage

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.4...v1.11.5)

## [v1.11.4](https://github.com/rocket-pool/smartnode/tree/v1.11.4) (2024-01-10)

### Changed

- Update Teku version

### Fixed

- `GetBnOpenPorts` fixed

### Removed

- `web3.storage` refs

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.3...v1.11.4)

## [v1.11.3](https://github.com/rocket-pool/smartnode/tree/v1.11.3) (2024-01-09)

### Changed

- Update Besu version
- Update Nimbus version
- Update Geth version
- Update Lodestar version

### Added

- Prompt if the user would prefer to skip key recovery when recovering wallet with unsynced clients (https://github.com/rocket-pool/smartnode/pull/425)
- Mainnet node sync check (https://github.com/rocket-pool/smartnode/issues/5)
- Treegen v8 (https://github.com/rocket-pool/smartnode/pull/424)

### Fixed

- SNG amnesiac restart-after-stake text (https://github.com/rocket-pool/smartnode/pull/426)
- Calculating graffiti (https://github.com/rocket-pool/smartnode/pull/428)

### Removed

- envsubst (https://github.com/rocket-pool/smartnode/pull/420)
- web3.storage dependency (https://github.com/rocket-pool/smartnode/pull/414)
- Nethermind complete history box

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.2...v1.11.3)

## [v1.11.2](https://github.com/rocket-pool/smartnode/tree/v1.11.2) (2023-12-19)

### Changed

- Update Teku version
- Update Nethermind version
- Continue submission when the file upload fails (https://github.com/rocket-pool/smartnode/pull/416)
- Dedup calls to LoadConfig (https://github.com/rocket-pool/smartnode/pull/418)

### Added

- Print rescue node plugin info even when the EC is syncing (https://github.com/rocket-pool/smartnode/pull/419)
- Scroll submit price logic (https://github.com/rocket-pool/smartnode/pull/422)
- Checks to ensure if all ports in the configuration are unique (https://github.com/rocket-pool/smartnode/pull/423)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.1...v1.11.2)

## [v1.11.1](https://github.com/rocket-pool/smartnode/tree/v1.11.1) (2023-12-05)

### Changed

- Client versions updated (https://github.com/rocket-pool/smartnode/pull/417)

### Added

- Prevented truncation of user-settings.yml when disk is full by writing to /tmp first and moving the result to user-settings.yml (https://github.com/rocket-pool/smartnode/pull/411)
- PBSS the default for new installations (https://github.com/rocket-pool/smartnode/pull/404)
- Add a Rescue Node add-on to make it easier for people to connect and disconnect (https://github.com/rocket-pool/smartnode/pull/402)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.11.0...v1.11.1)

## [v1.11.0](https://github.com/rocket-pool/smartnode/tree/v1.11.0) (2023-10-12)

### Changed

- Update Besu version
- Update Teku version
- Update Nethermind version
- Update Lodestar version
- Update Lighthouse version
- Update Nimbus version
- Update Prysm version
- Update Grafana
- Update Prometheus
- RPL node op inflation goes to the pDAO if no nodes are eligible
- Mainnet and Prater V7 intervals
- Cache size for Nethermind

### Added

- Holesky support
- PBSS
- Rewards file v2 and accompanying interfaces
- No pruning for Geth with PBSS

### Removed

- Cache for Geth
- Empty MPs from the index map during serialization
- Blocknative

### Fixed

- Non-staking MPs being added to the RR cache

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.10.2...v1.11.0)

## [v1.10.2](https://github.com/rocket-pool/smartnode/tree/v1.10.2) (2023-08-30)

### Changed

- Update Teku version
- Update Besu version
- Update Prometheus version
- Update Geth version
- Update Prysm version
- Update Nimbus version
- Minipool distribute now sorts by most-to-least rewards and has a threshold flag

### Added

- 200 idle HTTP connections
- Note to the daemons about the current mode

### Removed

- Bloxroute ethical and the non-sandwiching profile

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.10.1...v1.10.2)

## [v1.10.1](https://github.com/rocket-pool/smartnode/tree/v1.10.1) (2023-08-10)

### Changed

- Update Lodestar version
- Update Lighthouse version

### Added

- Arbitrary ERC20 support to node-send
- Safety checks and warnings to node-send
- Support for price submission to Base

### Fixed

- RR issue during subsequent intervals if the first block is missing
- Missing interval rollover check for oDAO members

### Removed

- Finalized validators from the staking count in Grafana
- Old flags from the daemon

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.10.0...v1.10.1)

## [v1.10.0](https://github.com/rocket-pool/smartnode/tree/v1.10.0) (2023-07-24)

### Changed

- "Expose API Port" from bool to a ternary choice (https://github.com/rocket-pool/smartnode/pull/342)
- BN client Committees response optimizations (https://github.com/rocket-pool/smartnode/pull/361)
- Slash timer messaging more explicit, ominous, and linked to the docs (https://github.com/rocket-pool/smartnode/pull/377)
- google.golang.org/grpc from 1.52.3 to 1.53.0 (https://github.com/rocket-pool/smartnode/pull/375)
- Change []beacon.Committee to an interface to reduce copying on the hot path
- Client ready checks deduplicated (https://github.com/rocket-pool/smartnode/pull/378)
- Way smartnode downloads and runs the install script to catch errors with the download better is changed (https://github.com/rocket-pool/smartnode/pull/381)
- Hoisted nonce validation, refactored to fluent for client init (https://github.com/rocket-pool/smartnode/pull/383)

### Removed

- Removed debug line (shared/services/rewards/rolling-record.go, line #197)

[Full Changelog](https://github.com/rocket-pool/smartnode/compare/v1.9.8...v1.10.0)
