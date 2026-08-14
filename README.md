<picture>
  <img alt="Rocket Pool - Decentralised Ethereum Staking Protocol" src="https://raw.githubusercontent.com/rocket-pool/.github/main/assets/logo.svg" width="auto" height="120">
</picture>

# Rocket Pool - Smart Node Package

Rocket Pool is a next generation Ethereum Proof-of-Stake (PoS) infrastructure service designed to be highly decentralised, distributed, and compatible with Ethereum's new consensus protocol.

Running a Rocket Pool Smart Node allows you to stake on Ethereum with only 4 ETH.
You can earn a higher return than you would outside the network by capturing up to an additional 14% commission on staked ETH as well as RPL rewards.

This repository contains the source code for:

- The Rocket Pool Smart Node client (CLI), which is used to manage a smart node either locally or remotely (over SSH)
- The Rocket Pool Smart Node service, which provides an API for client communication and performs background node tasks

The Smart Node service is designed to be run as part of a Docker stack and generally does not need to be installed manually.
See the [Rocket Pool dockerhub](https://hub.docker.com/u/rocketpool) page for a complete list of Docker images.

## Installation

See the [Smart Node Installer](https://github.com/rocket-pool/smartnode-install) repository for supported platforms and installation instructions.

## Development

A [Makefile](./Makefile) is included for building, testing, and linting.

- `make` or `make default` will build rocketpool-cli, rocketpool-daemon, treegen, and run the linter.
  - build/rocketpool-cli and build/rocketpool-daemon will be symlinked to the version and architecture specific binaries in build/
- `make all` will build rocketpool-cli, rocketpool-daemon, treegen, and run the linter.
  - symlinks will be created for the first 3 binaries in build/
- `make release` will build all architecture specific binaries as well as docker images and manifests
  - It will tag docker images as latest as well as the version in `shared/version.txt`
  - It will put cli and native mode binaries in build/\<version\>
- `make build/rocketpool-cli` builds just the cli
  - The build is done in docker, unless you run `make NO_DOCKER=true \<cmd\>`
- `make build/rocketpool-daemon` builds just the daemon
  - The build is done in docker, unless you run `make NO_DOCKER=true \<cmd\>`
- `make build/treegen` builds just the treegen stand-alone binary.
  - The build is done in docker, unless you run `make NO_DOCKER=true \<cmd\>`
- `make docker` builds the rocketpool/smartnode containers for all supported architectures and saves them in build/\<version\>/docker
- `make docker-push` builds, loads, pushes, creates a multi-arch manifest, and pushes the smartnode containers and manifest.
- `make docker-latest` does the same as docker-push, but tags latest which is also pushed.
- `make lint` runs the linter.
- `make test` runs all unit tests.
- `make clean` deletes any binaries. It does not clear your go caches. It does not clean up old docker images.

## CLI Commands

The following commands are available via the Smart Node client:

### COMMANDS:

- **auction**, a - Manage Rocket Pool RPL auctions
  - `rocketpool auction status, s` - Get RPL auction status
  - `rocketpool auction lots, l` - Get RPL lots for auction
  - `rocketpool auction create-lot, t` - Create a new lot
  - `rocketpool auction bid-lot, b` - Bid on a lot
  - `rocketpool auction claim-lot, c` - Claim RPL from a lot
  - `rocketpool auction recover-lot, r` - Recover unclaimed RPL from a lot (returning it to the auction contract)
- **claims**, l - View and claim all available rewards and credits across the node
  - `rocketpool claims status, s` - Display all available rewards and credits across the node without claiming
  - `rocketpool claims claim-all, c` - Display all available rewards and credits and claim them
- **minipool**, m - Manage the node's minipools
  - `rocketpool minipool status, s` - Get a list of the node's minipools
  - `rocketpool minipool stake, t` - Stake a minipool after the scrub check, moving it from prelaunch to staking.
  - `rocketpool minipool set-withdrawal-creds, swc` - Convert the withdrawal credentials for a migrated solo validator from the old 0x00 value to the minipool address. Required to complete the migration process.
  - `rocketpool minipool import-key, ik` - Import the externally-derived key for a minipool that was previously a solo validator, so the Smart Node's VC manages it instead of your externally-managed VC.
  - `rocketpool minipool refund, r` - Refund ETH belonging to the node from minipools
  - `rocketpool minipool distribute-balance, d` - Distribute a minipool's ETH balance between your withdrawal address and the rETH holders.
  - `rocketpool minipool exit, e` - Exit staking minipools from the beacon chain
  - `rocketpool minipool close, c` - Withdraw any remaining balance from a minipool and close it
  - `rocketpool minipool delegate-upgrade, u` - Upgrade a minipool's delegate contract to the latest version
  - `rocketpool minipool rescue-dissolved, rd` - Manually deposit ETH into the Beacon deposit contract for a dissolved minipool, activating it on the Beacon Chain so it can be exited.
- **megapool**, g - Manage the node's megapool
  - `rocketpool megapool deposit, d` - Make a deposit and create a new validator on the megapool. Optionally specify count to make multiple deposits.
  - `rocketpool megapool status, s` - Get the node's megapool status
  - `rocketpool megapool validators, v` - Get a list of the megapool's validators
  - `rocketpool megapool repay-debt, r` - Repay megapool debt
  - `rocketpool megapool reduce-bond, e` - Reduce the megapool bond
  - `rocketpool megapool claim, c` - Claim any megapool rewards that were distributed but not yet claimed
  - `rocketpool megapool stake, k` - Stake a megapool validator
  - `rocketpool megapool exit-queue, x` - Exit the megapool queue
  - `rocketpool megapool dissolve-validator, i` - Dissolve a megapool validator
  - `rocketpool megapool exit-validator, t` - Request to exit a megapool validator
  - `rocketpool megapool notify-validator-exit, n` - Notify that a validator exit is in progress
  - `rocketpool megapool notify-final-balance, f` - Notify that a validator exit has completed and the final balance has been withdrawn
  - `rocketpool megapool distribute, b` - Distribute any accrued execution layer rewards sent to this megapool
  - `rocketpool megapool delegate-upgrade, u` - Upgrade a megapool's delegate contract to the latest version
  - `rocketpool megapool set-use-latest-delegate, l` - Set the megapool to always use the latest delegate
- **network**, e - Manage Rocket Pool network parameters
  - `rocketpool network stats, s` - Get stats about the Rocket Pool network and its tokens
  - `rocketpool network timezone-map, t` - Shows a table of the timezones that node operators belong to
  - `rocketpool network node-fee, f` - Get the current network node commission rate
  - `rocketpool network rpl-price, p` - Get the current network RPL price in ETH
  - `rocketpool network generate-rewards-tree, g` - Generate and save the rewards tree file for the provided interval.
    Note that this is an asynchronous process, so it will return before the file is generated.
    You will need to use `rocketpool service logs api` to follow its progress.
  - `rocketpool network dao-proposals, d` - Get the currently active DAO proposals
- **node**, n - Manage the node
  - `rocketpool node status, s` - Get the node's status
  - `rocketpool node sync, y` - Get the sync progress of the eth1 and eth2 clients
  - `rocketpool node register, r` - Register the node with Rocket Pool
  - `rocketpool node rewards, e` - Get the time and your expected RPL rewards of the next checkpoint
  - `rocketpool node set-primary-withdrawal-address, w` - Set the node's primary withdrawal address, which will receive all ETH rewards (and RPL if the RPL withdrawal address is not set)
  - `rocketpool node confirm-primary-withdrawal-address, f` - Confirm the node's pending primary withdrawal address if it has been set back to the node's address itself
  - `rocketpool node set-rpl-withdrawal-address, srwa` - Set the node's RPL withdrawal address, which will receive all RPL rewards and staked RPL withdrawals
  - `rocketpool node confirm-rpl-withdrawal-address, crwa` - Confirm the node's pending rpl withdrawal address if it has been set back to the node's address itself
  - `rocketpool node allow-rpl-locking, arl` - Allow the node to lock RPL when creating governance proposals/challenges
  - `rocketpool node deny-rpl-locking, drl` - Do not allow the node to lock RPL when creating governance proposals/challenges
  - `rocketpool node set-timezone, t` - Set the node's timezone location
  - `rocketpool node swap-rpl, p` - Swap old RPL for new RPL
  - `rocketpool node stake-rpl, k` - Stake RPL against the node
  - `rocketpool node add-address-to-stake-rpl-whitelist, asw` - Adds an address to your node's RPL staking whitelist, so it can stake RPL on behalf of your node.
  - `rocketpool node remove-address-from-stake-rpl-whitelist, rsw` - Removes an address from your node's RPL staking whitelist, so it can no longer stake RPL on behalf of your node.
  - `rocketpool node claim-rewards, c` - Claim available RPL and ETH rewards for any checkpoint you haven't claimed yet
  - `rocketpool node withdraw-rpl, i` - Withdraw RPL staked against the node
  - `rocketpool node withdraw-eth, h` - Withdraw ETH staked on behalf of the node
  - `rocketpool node withdraw-credit, wc` - (Saturn) Withdraw ETH credit from the node as rETH
  - `rocketpool node send, n` - Send ETH or tokens from the node account to an address. ENS names supported. Use 'all' as the amount to send the entire balance. <token> can be 'rpl', 'eth', 'fsrpl' (for the old RPL v1 token), 'reth', or the address of an arbitrary token you want to send (including the 0x prefix).
  - `rocketpool node initialize-fee-distributor, z` - Create the fee distributor contract for your node, so you can withdraw priority fees and MEV rewards after the merge
  - `rocketpool node distribute-fees, b` - Distribute the priority fee and MEV rewards from your fee distributor to your withdrawal address and the rETH contract (based on your node's average commission)
  - `rocketpool node join-smoothing-pool, js` - Opt your node into the Smoothing Pool
  - `rocketpool node leave-smoothing-pool, ls` - Leave the Smoothing Pool
  - `rocketpool node sign-message, sm` - Sign an arbitrary message with the node's private key
  - `rocketpool node send-message` - Send a zero-ETH transaction to the target address (or ENS) with the provided hex-encoded message as the data payload
  - `rocketpool node claim-unclaimed-rewards, cur` - Sends any unclaimed rewards to the node's withdrawal address
  - `rocketpool node provision-express-tickets, pet` - Provision the node's express tickets
- **odao**, o - Manage the Rocket Pool oracle DAO
  - `rocketpool odao status, s` - Get oracle DAO status
  - `rocketpool odao members, m` - Get the oracle DAO members
  - `rocketpool odao member-settings, b` - Get the oracle DAO settings related to oracle DAO members
  - `rocketpool odao proposal-settings, a` - Get the oracle DAO settings related to oracle DAO proposals
  - `rocketpool odao minipool-settings, i` - Get the oracle DAO settings related to minipools
  - `rocketpool odao penalise-megapool, pm` - (Saturn) Penalise a megapool
  - `rocketpool odao propose, p` - Make an oracle DAO proposal
  - `rocketpool odao proposals, o` - Manage oracle DAO proposals
  - `rocketpool odao join, j` - Join the oracle DAO (requires an executed invite proposal)
  - `rocketpool odao leave, l` - Leave the oracle DAO (requires an executed leave proposal)
  - `rocketpool odao upgrade, u` - Upgrade proposals
- **pdao**, p - Manage the Rocket Pool Protocol DAO
  - `rocketpool pdao status, s` - Get the pdao status
  - `rocketpool pdao settings, st` - Show all of the current Protocol DAO settings and values
  - `rocketpool pdao rewards-percentages, rp` - View the RPL rewards allocation percentages for node operators, the Oracle DAO, and the Protocol DAO
  - `rocketpool pdao set-signalling-address, ssa` - Set the address you want to use to represent your node on Snapshot
  - `rocketpool pdao clear-signalling-address, csa` - Clear the node's signalling address
  - `rocketpool pdao set-voting-delegate, svd` - Set the address you want to use when voting on Rocket Pool on-chain governance proposals, or the address you want to delegate your voting power to.
  - `rocketpool pdao claim-bonds, cb` - Unlock any bonded RPL you have for a proposal or set of challenges, and claim any bond rewards for defending or defeating the proposal
  - `rocketpool pdao propose, p` - Make a Protocol DAO proposal
    - `rocketpool pdao propose submit-batch, sb` - Submit a single proposal that changes multiple Protocol DAO settings from a JSON file
    - Setting propose commands accept `--to-json <file>` to write/append a setting change to a JSON file instead of submitting a transaction
  - `rocketpool pdao proposals, o` - Manage Protocol DAO proposals
- **queue**, q - Manage the Rocket Pool deposit queue
  - `rocketpool queue status, s` - Get the deposit pool and minipool queue status
  - `rocketpool queue process, p` - Process the deposit pool
  - `rocketpool queue assign-deposits, ad` - Assign deposits to queued validators
- **security**, c - Manage the Rocket Pool security council
  - `rocketpool security status, s` - Get security council status
  - `rocketpool security members, m` - Get the security council members
  - `rocketpool security propose, p` - Make a security council proposal
  - `rocketpool security proposals, o` - Manage security council proposals
  - `rocketpool security join, j` - Join the security council (requires an executed invite proposal)
  - `rocketpool security leave, l` - Leave the security council (requires an executed leave proposal)
- **service**, s - Manage Rocket Pool service
  - `rocketpool service install, i` - Install the Rocket Pool service
  - `rocketpool service config, c` - Configure the Rocket Pool service
  - `rocketpool service status, u` - View the Rocket Pool service status
  - `rocketpool service start, s` - Start the Rocket Pool service
  - `rocketpool service stop, pause, p, o` - Pause the Rocket Pool service
  - `rocketpool service reset-docker, rd` - Cleanup Docker resources, including stopped containers, unused images and networks. Stops and restarts Smart Node.
  - `rocketpool service prune-docker, pd` - Cleanup unused Docker resources, including stopped containers, unused images, networks and volumes. Does not restart smartnode, so the running containers and the images and networks they reference will not be pruned.
  - `rocketpool service logs, l` - View the Rocket Pool service logs
  - `rocketpool service compose` - View the Rocket Pool service docker compose config
  - `rocketpool service version, v` - View the Rocket Pool service version information
  - `rocketpool service prune-eth1, n` - Shuts down the main ETH1 client and prunes its database, freeing up disk space, then restarts it when it's done.
  - `rocketpool service install-update-tracker, d` - Install the update tracker that provides the available system update count to the metrics dashboard
  - `rocketpool service get-config-yaml` - Generate YAML that shows the current configuration schema, including all of the parameters and their descriptions
  - `rocketpool service resync-eth1` - Deletes the main ETH1 client's chain data and resyncs it from scratch. Only use this as a last resort!
  - `rocketpool service resync-eth2` - Deletes the ETH2 client's chain data and resyncs it from scratch. Only use this as a last resort!
  - `rocketpool service terminate, t` - Deletes all of the Rocket Pool Docker containers and volumes, including your ETH1 and ETH2 chain data and your Prometheus database (if metrics are enabled). Also removes your entire `.rocketpool` configuration folder, including your wallet, password, and validator keys. Only use this if you are cleaning up the Smart Node and want to start over!
- **wallet**, w - Manage the node wallet
  - `rocketpool wallet status, s` - Get the node wallet status
  - `rocketpool wallet init, i` - Initialize the node wallet
  - `rocketpool wallet recover, r` - Recover a node wallet from a mnemonic phrase
  - `rocketpool wallet rebuild, b` - Rebuild validator keystores from derived keys
  - `rocketpool wallet test-recovery, t` - Test recovering a node wallet without actually generating any of the node wallet or validator key files to ensure the process works as expected
  - `rocketpool wallet export, e` - Export the node wallet in JSON format
  - `rocketpool wallet set-ens-name, ens` - Set a name to the node wallet's ENS reverse record
  - `rocketpool wallet purge` - Deletes your node wallet, your validator keys, and restarts your Validator Client while preserving your chain data. WARNING: Only use this if you want to stop validating with this machine!
  - `rocketpool wallet masquerade, m` - Change your node's effective address to a different one. Your node will not be able to submit transactions or sign messages since you don't have the corresponding wallet's private key.
  - `rocketpool wallet end-masquerade, em` - End a masquerade, restoring your node's effective address back to your wallet address if one is loaded.
- **update** - Update the cli binary
- **help**, h - Shows a list of commands or help for one command

### GLOBAL OPTIONS:

- `rocketpool --allow-root, -r` - Allow rocketpool to be run as the root user
- `rocketpool --config-path path, -c path` - Rocket Pool config asset path (default: "~/.rocketpool")
- `rocketpool --daemon-path path, -d path` - Interact with a Rocket Pool service daemon at a path on the host OS, running outside of docker
- `rocketpool --maxFee value, -f value` - The max fee (including the priority fee) you want a transaction to cost, in gwei (default: 0)
- `rocketpool --maxPrioFee value, -i value` - The max priority fee you want a transaction to use, in gwei (default: 0)
- `rocketpool --gasLimit value, -l value` - [DEPRECATED] Desired gas limit (default: 0)
- `rocketpool --nonce value, -n value` - Use this flag to explicitly specify the nonce that this transaction should use, so it can override an existing 'stuck' transaction
- `rocketpool --debug` - Enable debug printing of API commands
- `rocketpool --secure-session, -s` - Some commands may print sensitive information to your terminal. Use this flag when nobody can see your screen to allow sensitive data to be printed without prompting
- `rocketpool --help, -h` - show help
- `rocketpool --version, -v` - print the version
