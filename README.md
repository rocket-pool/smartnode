# Rocket Pool Rewards Tree Generation Tool

This is a standalone tool for creating the rewards tree and minipool attestation files for rewards intervals on the Rocket Pool network.
It can recreate trees for past rewards intervals, or it can "simulate" the tree for the current interval ending at the latest finalized block (for testing purposes).


## Building

To build it, simply enter this folder and run `go build`.
You will need to have Go v1.19 set up already.
