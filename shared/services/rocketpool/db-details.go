package rocketpool

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"
)

// GethDBVersion reads metadata filenames without opening or locking the DB.
// Smart Node explicitly sets --datadir /ethclient/geth on every network.
func (c *Client) GethDBVersion(executionContainerName string) (string, error) {
	const inspect = `if [ ! -d /ethclient/geth/geth/chaindata ]; then exit 0; fi; ls -1 /ethclient/geth/geth/chaindata`
	output, err := c.readOutput(fmt.Sprintf("docker exec %s sh -c %s", shellescape.Quote(executionContainerName), shellescape.Quote(inspect)))
	if err != nil {
		return "", fmt.Errorf("could not inspect Geth database (the execution container must be running): %w", err)
	}
	return gethDBVersionFromFiles(string(output))
}

func gethDBVersionFromFiles(listing string) (string, error) {
	var hasCurrent, hasManifest, hasOptions, hasFormat bool
	var latestGeneration uint64
	format := uint64(1) // Pebble's legacy FormatMostCompatible has no format marker.
	for _, name := range strings.Split(strings.TrimSpace(listing), "\n") {
		switch {
		case name == "CURRENT":
			hasCurrent = true
		case strings.HasPrefix(name, "marker.manifest."):
			hasManifest = true
		case strings.HasPrefix(name, "OPTIONS"):
			hasOptions = true
		case strings.HasPrefix(name, "marker.format-version."):
			// Atomic markers encode marker.<name>.<generation>.<value>.
			// Crashes may leave older markers behind; use the latest generation.
			parts := strings.Split(name, ".")
			if len(parts) != 4 {
				return "", fmt.Errorf("invalid Pebble format marker %q", name)
			}
			generation, err := strconv.ParseUint(parts[2], 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid Pebble format marker %q: %w", name, err)
			}
			version, err := strconv.ParseUint(parts[3], 10, 64)
			if err != nil || version == 0 {
				return "", fmt.Errorf("invalid Pebble format version in %q", name)
			}
			if !hasFormat || generation > latestGeneration {
				latestGeneration, format = generation, version
			}
			hasFormat = true
		}
	}
	// Match Geth core/rawdb.PreexistingDatabase, including old LevelDB nodes.
	if !hasManifest && !(hasCurrent && hasOptions) {
		if hasCurrent {
			return "leveldb", nil
		}
		return "none", nil
	}
	// Geth ethdb/pebble/version.go uses FormatFlushableIngest (13) as the
	// minimum on-disk format opened with Pebble v2, including migrated DBs.
	if format >= 13 {
		return "v2", nil
	}
	return "v1", nil
}
