# Rocket Pool Rewards Tree Generation Tool

This is a standalone tool for creating the rewards tree and minipool attestation files for rewards intervals on the Rocket Pool network.
It can recreate trees for past rewards intervals, or it can "simulate" the tree for the current interval ending at the latest finalized block (for testing purposes).
It uses the same codebase as the Smartnode, so you can be assured that `treegen` will generate the same trees as the Smartnode stack. 


## Running Treegen

There are currently three ways to run `treegen`:

1. Run the precompiled binaries locally (Linux only, using `glibc`)
2. Run the Docker image (Linux, Windows, and macOS)
3. Build from source and run locally


### Running the Binary Locally

Run the binary as follows:

```
$ ./treegen-linux-amd64 [options]
```

Options:

```
   --bn-endpoint value, -b value  The URL of the Beacon Node's REST API. Note that for past interval generation, this must have Archive capability (ability to replay arbitrary historical states). (default: "http://localhost:5052")
   --ec-endpoint value, -e value  The URL of the Execution Client's JSON-RPC API. Note that for past interval generation, this must be an Archive EC. (default: "http://localhost:8545")
   --interval value, -i value     The rewards interval to generate the artifacts for. A value of -1 indicates that you want to do a "dry run" of generating the tree for the current (active) interval, using the current latest finalized block as the interval end. (default: -1)
   --output-dir value, -o value   Optional output directory to save generated files (default is the current working directory).
   --pretty-print, -p             Toggle for saving the files in pretty-print format so they're human readable. (default: true)
   --ruleset value, -r value      The ruleset to use during generation. If not included, treegen will use the default ruleset for the network based on the rewards interval at the chosen block. Default of 0 will use whatever the ruleset specified by the network based on which block is being targeted. (default: 0)
   --network-info, -n             If provided, this will simply print out info about the network being used, the current rewards interval, and the current ruleset. (default: false)
   --approximate-only, -a         Approximates the rETH stakers' share of the Smoothing Pool at the current block instead of generating the entire rewards tree. Ignores -i. (default: false)
   --use-rolling-records, -rr     Enable the rolling record capability of the Smartnode tree generator. Use this to store and load record caches instead of recalculating attestation performance each time you run treegen. (default: false)
```


### Running via the Docker Image

`treegen` is also available in a Docker image for users not on Linux systems, or on systems incompatible with the precompiled binaries.
To run it, use the following command:

```
$ ./treegen.sh -e <EC endpoint> -b <BN endpoint> [options]
```

Options:

```
   --bn-endpoint value, -b value  The URL of the Beacon Node's REST API. Note that for past interval generation, this must have Archive capability (ability to replay arbitrary historical states). (default: "http://localhost:5052")
   --ec-endpoint value, -e value  The URL of the Execution Client's JSON-RPC API. Note that for past interval generation, this must be an Archive EC. (default: "http://localhost:8545")
   --interval value, -i value     The rewards interval to generate the artifacts for. A value of -1 indicates that you want to do a "dry run" of generating the tree for the current (active) interval, using the current latest finalized block as the interval end. (default: -1)
   --pretty-print, -p             Toggle for saving the files in pretty-print format so they're human readable. (default: true)
   --ruleset value, -r value      The ruleset to use during generation. If not included, treegen will use the default ruleset for the network based on the rewards interval at the chosen block. Default of 0 will use whatever the ruleset specified by the network based on which block is being targeted. (default: 0)
   --network-info, -n             If provided, this will simply print out info about the network being used, the current rewards interval, and the current ruleset. (default: false)
   --approximate-only, -a         Approximates the rETH stakers' share of the Smoothing Pool at the current block instead of generating the entire rewards tree. Ignores -i. (default: false)
   --use-rolling-records, -rr     Enable the rolling record capability of the Smartnode tree generator. Use this to store and load record caches instead of recalculating attestation performance each time you run treegen. (default: false)
```

NOTE: Do *not* use the `-o` flag if you are using this script, as it is already built into the script.
Output files will be stored in the `out` directory.


## Building

To build the binary locally, simply enter this folder and run `go build`.
You will need to have Go v1.19 or higher set up already.
