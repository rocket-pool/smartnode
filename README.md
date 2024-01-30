# Rocket Pool - Smart Node Package

<p align="center">
  <img src="https://raw.githubusercontent.com/rocket-pool/rocketpool/master/images/logo.png?raw=true" alt="Rocket Pool - Next Generation Decentralised Ethereum Proof-of-Stake (PoS) Infrastructure Service and Pool" width="500" />
</p>

---

Rocket Pool is a next generation Ethereum Proof-of-Stake (PoS) infrastructure service designed to be highly decentralised, distributed, and compatible with Ethereum's new consensus protocol.

Running a Rocket Pool smart node allows you to stake on Ethereum with only 16 ETH and 1.6 ETH worth of Rocket Pool's RPL token.
You can earn a higher return than you would outside the network by capturing an additional 15% commission on staked ETH as well as RPL rewards.

This repository contains the source code for:

* The Rocket Pool Smartnode client (CLI), which is used to manage a smart node either locally or remotely (over SSH)
* The Rocket Pool Smartnode service, which provides an API for client communication and performs background node tasks

The Smartnode service is designed to be run as part of a Docker stack and generally does not need to be installed manually.
See the [Rocket Pool dockerhub](https://hub.docker.com/u/rocketpool) page for a complete list of Docker images.


## Installation

See the [Smartnode Installer](https://github.com/rocket-pool/smartnode-install) repository for supported platforms and installation instructions.


## CLI Commands

The following commands are available via the Smartnode client:


### COMMANDS:
- **auction**, a - Manage Rocket Pool RPL auctions
  - `rocketpool auction status, s` - Get RPL auction status
  - `rocketpool auction lots, l` - Get RPL lots for auction
  - `rocketpool auction create-lot, t` - Create a new lot
  - `rocketpool auction bid-lot, b` - Bid on a lot
  - `rocketpool auction claim-lot, c` - Claim RPL from a lot
  - `rocketpool auction recover-lot, r` - Recover unclaimed RPL from a lot (returning it to the auction contract)
- **minipool**, m - Manage the node's minipools
  - `rocketpool minipool status, s` - Get a list of the node's minipools
  - `rocketpool minipool stake, t` - Stake a minipool after the scrub check, moving it from prelaunch to staking.
  - `rocketpool minipool refund, r` - Refund ETH belonging to the node from minipools
  - `rocketpool minipool exit, e` - Exit staking minipools from the beacon chain
  - `rocketpool minipool delegate-upgrade, u` - Upgrade a minipool's delegate contract to the latest version
  - `rocketpool minipool delegate-rollback, b` - Roll a minipool's delegate contract back to its previous version
  - `rocketpool minipool set-use-latest-delegate, l` - If enabled, the minipool will ignore its current delegate contract and always use whatever the latest delegate is
  - `rocketpool minipool find-vanity-address, v` - Search for a custom vanity minipool address
- **network**, e - Manage Rocket Pool network parameters
  - `rocketpool network stats, s` - Get stats about the Rocket Pool network and its tokens
  - `rocketpool network timezone-map, t` - Shows a table of the timezones that node operators belong to
  - `rocketpool network node-fee, f` - Get the current network node commission rate
  - `rocketpool network rpl-price, p` - Get the current network RPL price in ETH
  - `rocketpool network generate-rewards-tree, g` - Generate and save the rewards tree file for the provided interval.
  Note that this is an asynchronous process, so it will return before the file is generated.
  You will need to use `rocketpool service logs api` to follow its progress.
  - `rocketpool network dao-proposals, d` - Get the currently active DAO proposals
  - `rocketpool network initialize-voting, iv` - Unlocks a node operator's voting power (only required for node operators who registered before governance structure was in place)
  - `rocketpool network set-voting-delegate, sod` - Delegates the node voting power to another address (for on-chain governance)
- **node**, n - Manage the node
  - `rocketpool node status, s` - Get the node's status
  - `rocketpool node sync, y` - Get the sync progress of the eth1 and eth2 clients
  - `rocketpool node register, r` - Register the node with Rocket Pool
  - `rocketpool node rewards, e` - Get the time and your expected RPL rewards of the next checkpoint
  - `rocketpool node set-withdrawal-address, w` - Set the node's withdrawal address
  - `rocketpool node confirm-withdrawal-address, f` - Confirm the node's pending withdrawal address if it has been set back to the node's address itself
  - `rocketpool node set-timezone, t` - Set the node's timezone location
  - `rocketpool node swap-rpl, p` - Swap old RPL for new RPL
  - `rocketpool node stake-rpl, k` - Stake RPL against the node
  - `rocketpool node claim-rewards, c` - Claim available RPL and ETH rewards for any checkpoint you haven't claimed yet
  - `rocketpool node withdraw-rpl, i` - Withdraw RPL staked against the node
  - `rocketpool node deposit, d` - Make a deposit and create a minipool
  - `rocketpool node send, n` - Send ETH or tokens from the node account to an address
  - `rocketpool node set-voting-delegate, sv` - Set the address you want to use when voting on Rocket Pool governance proposals on Snapshot, or the address you want to delegate your voting power to.
  - `rocketpool node clear-voting-delegate, cv` - Remove the address you've set for voting on Rocket Pool governance proposals on Snapshot.
  - `rocketpool node initialize-fee-distributor, z` - Create the fee distributor contract for your node, so you can withdraw priority fees and MEV rewards after the merge
  - `rocketpool node distribute-fees, b` - Distribute the priority fee and MEV rewards from your fee distributor to your withdrawal address and the rETH contract (based on your node's average commission` -
  - `rocketpool node join-smoothing-pool, js` - Opt your node into the Smoothing Pool
  - `rocketpool node leave-smoothing-pool, ls` - Leave the Smoothing Pool
  - `rocketpool node sign-message, sm` - Sign an arbitrary message with the node's private key
- **odao**, o - Manage the Rocket Pool oracle DAO
  - `rocketpool odao status, s` - Get oracle DAO status
  - `rocketpool odao members, m` - Get the oracle DAO members
  - `rocketpool odao member-settings, b` - Get the oracle DAO settings related to oracle DAO members
  - `rocketpool odao proposal-settings, a` - Get the oracle DAO settings related to oracle DAO proposals
  - `rocketpool odao minipool-settings, i` - Get the oracle DAO settings related to minipools
  - `rocketpool odao propose, p` - Make an oracle DAO proposal
  - `rocketpool odao proposals, o` - Manage oracle DAO proposals
  - `rocketpool odao join, j` - Join the oracle DAO (requires an executed invite proposal)
  - `rocketpool odao leave, l` - Leave the oracle DAO (requires an executed leave proposal)
- **queue**, q - Manage the Rocket Pool deposit queue
  - `rocketpool queue status, s` - Get the deposit pool and minipool queue status
  - `rocketpool queue process, p` - Process the deposit pool
- **service**, s - Manage Rocket Pool service
  - `rocketpool service install, i` - Install the Rocket Pool service
  - `rocketpool service config, c` - Configure the Rocket Pool service
  - `rocketpool service status, u` - View the Rocket Pool service status
  - `rocketpool service start, s` -  Start the Rocket Pool service
  - `rocketpool service pause, p` -  Pause the Rocket Pool service
  - `rocketpool service stop, o` - Pause the Rocket Pool service (alias of 'rocketpool service pause')
  - `rocketpool service logs, l` - View the Rocket Pool service logs
  - `rocketpool service stats, a` - View the Rocket Pool service stats
  - `rocketpool service compose` - View the Rocket Pool service docker compose config
  - `rocketpool service version, v` - View the Rocket Pool service version information
  - `rocketpool service prune-eth1, n` - Shuts down the main ETH1 client and prunes its database, freeing up disk space, then restarts it when it's done.
  - `rocketpool service install-update-tracker, d` - Install the update tracker that provides the available system update count to the metrics dashboard
  - `rocketpool service get-config-yaml` - Generate YAML that shows the current configuration schema, including all of the parameters and their descriptions
  - `rocketpool service export-eth1-data` - Exports the execution client (eth1) chain data to an external folder. Use this if you want to back up your chain data before switching execution clients.
  - `rocketpool service import-eth1-data` - Imports execution client (eth1) chain data from an external folder. Use this if you want to restore the data from an execution client that you previously backed up.
  - `rocketpool service resync-eth1` - Deletes the main ETH1 client's chain data and resyncs it from scratch. Only use this as a last resort!
  - `rocketpool service resync-eth2` - Deletes the ETH2 client's chain data and resyncs it from scratch. Only use this as a last resort!
  - `rocketpool service terminate, t` - Deletes all of the Rocket Pool Docker containers and volumes, including your ETH1 and ETH2 chain data and your Prometheus database (if metrics are enabled). Only use this if you are cleaning up the Smartnode and want to start over!
- **wallet**, w - Manage the node wallet
  - `rocketpool wallet status, s` - Get the node wallet status
  - `rocketpool wallet init, i` - Initialize the node wallet
  - `rocketpool wallet recover, r` - Recover a node wallet from a mnemonic phrase
  - `rocketpool wallet rebuild, b` - Rebuild validator keystores from derived keys
  - `rocketpool wallet test-recovery, t` - Test recovering a node wallet without actually generating any of the node wallet or validator key files to ensure the process works as expected
  - `rocketpool wallet export, e` - Export the node wallet in JSON format
  - `rocketpool wallet purge` - Deletes your node wallet, your validator keys, and restarts your Validator Client while preserving your chain data. WARNING: Only use this if you want to stop validating with this machine!
  - `rocketpool wallet set-ens-name` - Send a transaction from the node wallet to configure it's ENS name
- **help**, h - Shows a list of commands or help for one command


### GLOBAL OPTIONS:
 - `rocketpool --allow-root, -r` - Allow rocketpool to be run as the root user
 - `rocketpool --config-path path, -c path` - Rocket Pool config asset path (default: "~/.rocketpool")
 - `rocketpool --daemon-path path, -d path` - Interact with a Rocket Pool service daemon at a path on the host OS, running outside of docker
 - `rocketpool --maxFee value, -f value` - The max fee (including the priority fee) you want a transaction to cost, in gwei (default: 0)
 - `rocketpool --maxPrioFee value, -i value` - The max priority fee you want a transaction to use, in gwei (default: 0)
 - `rocketpool --gasLimit value, -l value` - [DEPRECATED] Desired gas limit (default: 0)
 - `rocketpool --nonce value` - Use this flag to explicitly specify the nonce that this transaction should use, so it can override an existing 'stuck' transaction
 - `rocketpool --debug` - Enable debug printing of API commands
 - `rocketpool --secure-session, -s` - Some commands may print sensitive information to your terminal. Use this flag when nobody can see your screen to allow sensitive data to be printed without prompting
 - `rocketpool --help, -h` - show help
 - `rocketpool --version, -v` - print the version
