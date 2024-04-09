package minipool

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
)

const (
	vanityPrefixFlag  string = "prefix"
	vanitySaltFlag    string = "salt"
	vanityThreadsFlag string = "threads"
	vanityAddressFlag string = "node-address"
)

func findVanitySalt(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the target prefix
	prefix := c.String(vanityPrefixFlag)
	if prefix == "" {
		prefix = utils.Prompt("Please specify the address prefix you would like to search for (must start with 0x):", "^0x[0-9a-fA-F]+$", "Invalid hex string")
	}
	if !strings.HasPrefix(prefix, "0x") {
		return fmt.Errorf("Prefix must start with 0x.")
	}
	targetPrefix, success := big.NewInt(0).SetString(prefix, 0)
	if !success {
		return fmt.Errorf("Invalid prefix: %s", prefix)
	}

	// Get the starting salt
	saltString := c.String(vanitySaltFlag)
	var salt *big.Int
	if saltString == "" {
		salt = big.NewInt(0)
	} else {
		salt, success = big.NewInt(0).SetString(saltString, 0)
		if !success {
			return fmt.Errorf("Invalid starting salt: %s", salt)
		}
	}

	// Get the core count
	threads := c.Int(vanityThreadsFlag)
	if threads == 0 {
		threads = runtime.GOMAXPROCS(0)
	} else if threads < 0 {
		threads = 1
	} else if threads > runtime.GOMAXPROCS(0) {
		threads = runtime.GOMAXPROCS(0)
	}

	// Get the node address
	nodeAddressStr := c.String(vanityAddressFlag)
	if nodeAddressStr == "" {
		nodeAddressStr = "0"
	}

	// Get the vanity generation artifacts
	vanityArtifacts, err := rp.Api.Minipool.GetVanityArtifacts(nodeAddressStr)
	if err != nil {
		return err
	}

	// Set up some variables
	nodeAddress := vanityArtifacts.Data.NodeAddress.Bytes()
	minipoolFactoryAddress := vanityArtifacts.Data.MinipoolFactoryAddress
	initHash := vanityArtifacts.Data.InitHash.Bytes()
	shiftAmount := uint(42 - len(prefix))

	// Run the search
	fmt.Printf("Running with %d threads.\n", threads)

	wg := new(sync.WaitGroup)
	wg.Add(threads)
	stop := false
	stopPtr := &stop

	// Spawn worker threads
	start := time.Now()
	for i := 0; i < threads; i++ {
		saltOffset := big.NewInt(int64(i))
		workerSalt := big.NewInt(0).Add(salt, saltOffset)

		go func(i int) {
			foundSalt, foundAddress := runWorker(i == 0, stopPtr, targetPrefix, nodeAddress, minipoolFactoryAddress, initHash, workerSalt, int64(threads), shiftAmount)
			if foundSalt != nil {
				fmt.Printf("Found on thread %d: salt 0x%x = %s\n", i, foundSalt, foundAddress.Hex())
				*stopPtr = true
			}
			wg.Done()
		}(i)
	}

	// Wait for the workers to finish and print the elapsed time
	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)
	fmt.Printf("Finished in %s\n", elapsed)

	// Return
	return nil

}

func runWorker(report bool, stop *bool, targetPrefix *big.Int, nodeAddress []byte, minipoolManagerAddress common.Address, initHash []byte, salt *big.Int, increment int64, shiftAmount uint) (*big.Int, common.Address) {
	saltBytes := [32]byte{}
	hashInt := big.NewInt(0)
	incrementInt := big.NewInt(increment)
	hasher := crypto.NewKeccakState()
	nodeSalt := common.Hash{}
	addressResult := common.Hash{}

	// Set up the reporting ticker if requested
	var ticker *time.Ticker
	var tickerChan chan struct{}
	lastSalt := big.NewInt(0).Set(salt)
	if report {
		start := time.Now()
		reportInterval := 5 * time.Second
		ticker = time.NewTicker(reportInterval)
		tickerChan = make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					delta := big.NewInt(0).Sub(salt, lastSalt)
					deltaFloat, suffix := humanize.ComputeSI(float64(delta.Uint64()) / 5.0)
					deltaString := humanize.FtoaWithDigits(deltaFloat, 2) + suffix
					fmt.Printf("At salt 0x%x... %s (%s salts/sec)\n", salt, time.Since(start), deltaString)
					lastSalt.Set(salt)
				case <-tickerChan:
					ticker.Stop()
					return
				}
			}
		}()
	}

	// Run the main salt finder loop
	for {
		if *stop {
			return nil, common.Address{}
		}

		// Some speed optimizations -
		// This block is the fast way to do `nodeSalt := crypto.Keccak256Hash(nodeAddress, saltBytes)`
		salt.FillBytes(saltBytes[:])
		hasher.Write(nodeAddress)
		hasher.Write(saltBytes[:])
		hasher.Read(nodeSalt[:])
		hasher.Reset()

		// This block is the fast way to do `crypto.CreateAddress2(minipoolManagerAddress, nodeSalt, initHash)`
		// except instead of capturing the returned value as an address, we keep it as bytes. The first 12 bytes
		// are ignored, since they are not part of the resulting address.
		//
		// Because we didn't call CreateAddress2 here, we have to call common.BytesToAddress below, but we can
		// postpone that until we find the correct salt.
		hasher.Write([]byte{0xff})
		hasher.Write(minipoolManagerAddress.Bytes())
		hasher.Write(nodeSalt[:])
		hasher.Write(initHash)
		hasher.Read(addressResult[:])
		hasher.Reset()

		hashInt.SetBytes(addressResult[12:])
		hashInt.Rsh(hashInt, shiftAmount*4)
		if hashInt.Cmp(targetPrefix) == 0 {
			if report {
				close(tickerChan)
			}
			address := common.BytesToAddress(addressResult[12:])
			return salt, address
		}
		salt.Add(salt, incrementInt)
	}
}
