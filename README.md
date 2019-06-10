TODO: Add a complete list of CLI commands when available and detailed descriptions + link to docs. Add more detailed descriptions of the requirements, installation guide and a FAQ. Also include a list of benefits for running a Rocket Pool smart node.

# Rocket Pool - A next generation Ethereum proof of stake (PoS) infrastructure service and pool

<p align="center">
  <img src="https://raw.githubusercontent.com/rocket-pool/rocketpool/master/images/logo.png?raw=true" alt="Rocket Pool - Next Generation Decentralised Ethereum Proof-of-Stake (PoS) Infrastructure Service and Pool" width="500" />
</p>

---

`Rocket Pool 2.0` is a next generation Ethereum proof of stake (PoS) infrastructure service designed to be highly decentralised, distributed and compatible with Casper 2.0, the new consensus protocol that Ethereum will transition to in late 2019.

This repository contains the Smart Node package required to run a Smart Node in the Rocket Pool network and earn a higher return staking ether than you would outside the network.

# Package

The Smart Node package allows a node operator to install all dependencies and begin running a full Rocket Pool Smart Node easily.

## CLI

The package contains a command line interface that will allow node operators to stake their own ether easily, monitor their nodes status and connect to the Rocket Pool network.

## Daemons

The package also contains several services that operate in the background called daemons. They monitor the beacon chain on the Ethereum network and also monitor the Rocket Pool network.

# CLI Commands

Rocket Pool Smart Nodes are primarily managed by the `rocketpool` CLI application. The following commands are available:

- `rocketpool node status` - Displays information about the status of the smart node
- `rocketpool node init` - Initialises the smart node with an account used for all transactions with the Rocket Pool network
- `rocketpool node register` - Registers the smart node with the Rocket Pool network
- `rocketpool node withdraw` - Withdraws resources from the smart node contract back to the node account
- `rocketpool node timezone` - Set's the smart node's timezone information for display on the Rocket Pool website

- `rocketpool deposit status` - Displays information about the smart node's current pending deposit, if any
- `rocketpool deposit required` - Displays the required deposit amount, RPL requirement and RPL ratio for the specified staking duration
- `rocketpool deposit reserve` - Reserves a deposit with the Rocket Pool network and calculates the ETH and RPL requirements to finalize it
- `rocketpool deposit cancel` - Cancels the smart node's current pending deposit
- `rocketpool deposit complete` - Completes the smart node's current pending deposit, sending any required ETH and RPL, and displays information about the created minipool

- `rocketpool minipool status` - Displays information about the node's current minipools
- `rocketpool minipool withdraw` - Withdraws the node's deposit from an initialized, withdrawn or timed out minipool

- `rocketpool fee display` - Displays the current user fee charged by all node operators in the Rocket Pool network, and the target fee to vote for, if set locally
- `rocketpool fee set` - Sets the target user fee to vote for during node checkin, locally
