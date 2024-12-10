package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// A basic CLI tool which can be used to serialize NetworkState objects to files
// for future use.
// Accepts arguments for a beacon node URL, an execution node URL, and a slot number
// to get the state for.

var bnFlag = flag.String("b", "http://localhost:5052", "The beacon node URL")
var elFlag = flag.String("e", "http://localhost:8545", "The execution node URL")
var slotFlag = flag.Uint64("slot", 0, "The slot number to get the state for")
var networkFlag = flag.String("network", "mainnet", "The network to get the state for, i.e. 'mainnet' or 'holesky'")
var prettyFlag = flag.Bool("p", false, "Pretty print the output")
var inputFlag = flag.Bool("i", false, "Parse a network state from stdin instead of retrieving it from the network")
var criticalDutiesSlotsFlag = flag.Bool("critical-duties-slots", false, "If passed, output a list of critical duties slots for the given state as if it were the final state in a 6300 epoch interval. This is outputted instead of the state json.")
var criticalDutiesEpochCountFlag = flag.Uint64("critical-duties-epoch-count", 6300, "The number of epochs to consider when calculating critical duties")

func main() {
	flag.Parse()

	sn := config.NewSmartnodeConfig(nil)
	switch *networkFlag {
	case "mainnet":
		sn.Network.Value = cfgtypes.Network_Mainnet
	case "holesky":
		sn.Network.Value = cfgtypes.Network_Holesky
	default:
		fmt.Fprintf(os.Stderr, "Invalid network: %s\n", *networkFlag)
		fmt.Fprintf(os.Stderr, "Valid networks are: mainnet, holesky\n")
		os.Exit(1)
	}

	ec, err := ethclient.Dial(*elFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to execution node: %v\n", err)
		os.Exit(1)
	}

	contracts := sn.GetStateManagerContracts()
	fmt.Fprintf(os.Stderr, "Contracts: %+v\n", contracts)

	rocketStorage := sn.GetStorageAddress()

	rp, err := rocketpool.NewRocketPool(ec, common.HexToAddress(rocketStorage))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating rocketpool: %v\n", err)
		os.Exit(1)
	}
	bc := client.NewStandardHttpClient(*bnFlag)
	sm := state.NewNetworkStateManager(rp, contracts, bc, nil)

	var networkState *state.NetworkState

	if *inputFlag {
		decoder := json.NewDecoder(os.Stdin)
		err := decoder.Decode(&networkState)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding network state: %v\n", err)
			os.Exit(1)
		}
	} else if *slotFlag == 0 {
		fmt.Fprintf(os.Stderr, "Slot number not provided, defaulting to head slot.\n")
		networkState, err = sm.GetHeadState()
	} else {
		networkState, err = sm.GetStateForSlot(*slotFlag)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting network state: %v\n", err)
		os.Exit(1)
	}

	if *criticalDutiesSlotsFlag {
		criticalDutiesEpochs := state.NewCriticalDutiesEpochs(*criticalDutiesEpochCountFlag, networkState)
		fmt.Fprintf(os.Stderr, "Critical duties epochs to check: %d\n", len(criticalDutiesEpochs.CriticalDuties))

		criticalDutiesSlots, err := state.NewCriticalDutiesSlots(criticalDutiesEpochs, bc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating critical duties slots: %v\n", err)
			os.Exit(1)
		}

		// Serialize the critical duties slots to stdout
		encoder := json.NewEncoder(os.Stdout)
		if *prettyFlag {
			encoder.SetIndent("", "  ")
		}
		err = encoder.Encode(criticalDutiesSlots)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding critical duties slots: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
		return
	}

	fmt.Fprintf(os.Stderr, "Network state fetched, outputting to stdout\n")
	encoder := json.NewEncoder(os.Stdout)
	if *prettyFlag {
		encoder.SetIndent("", "  ")
	}
	err = encoder.Encode(networkState)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding network state: %v\n", err)
		os.Exit(1)
	}
}
