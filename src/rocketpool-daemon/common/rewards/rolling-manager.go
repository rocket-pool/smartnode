package rewards

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	recordsFilenameFormat         string = "%d-%d.json.zst"
	recordsFilenamePattern        string = "(?P<slot>\\d+)\\-(?P<epoch>\\d+)\\.json\\.zst"
	latestCompatibleVersionString string = "1.11.0-dev"
)

// Manager for RollingRecords
type RollingRecordManager struct {
	Record                       *RollingRecord
	LatestFinalizedEpoch         uint64
	ExpectedBalancesBlock        uint64
	ExpectedRewardsIntervalBlock uint64

	logger          *slog.Logger
	cfg             *config.SmartNodeConfig
	rp              *rocketpool.RocketPool
	bc              beacon.IBeaconClient
	mgr             *state.NetworkStateManager
	startSlot       uint64
	nextEpochToSave uint64

	beaconCfg            beacon.Eth2Config
	genesisTime          time.Time
	compressor           *zstd.Encoder
	decompressor         *zstd.Decoder
	recordsFilenameRegex *regexp.Regexp
}

// Creates a new manager for rolling records.
func NewRollingRecordManager(logger *slog.Logger, cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, bc beacon.IBeaconClient, mgr *state.NetworkStateManager, startSlot uint64, beaconCfg beacon.Eth2Config, rewardsInterval uint64) (*RollingRecordManager, error) {
	// Get the Beacon genesis time
	genesisTime := time.Unix(int64(beaconCfg.GenesisTime), 0)

	// Create the zstd compressor and decompressor
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, fmt.Errorf("error creating zstd compressor for rolling record manager: %w", err)
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd decompressor for rolling record manager: %w", err)
	}

	// Create the records filename regex
	recordsFilenameRegex := regexp.MustCompile(recordsFilenamePattern)

	// Make the records folder if it doesn't exist
	recordsPath := cfg.GetRecordsPath()
	fileInfo, err := os.Stat(recordsPath)
	if os.IsNotExist(err) {
		err2 := os.MkdirAll(recordsPath, 0755)
		if err2 != nil {
			return nil, fmt.Errorf("error creating rolling records folder: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking rolling records folder: %w", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("rolling records folder location exists (%s), but is not a folder", recordsPath)
	}

	sublogger := logger.With(slog.String(keys.RoutineKey, "Rolling Record"))
	logger.Info("Created Rolling Record manager.", slog.Uint64(keys.StartSlotKey, startSlot))
	return &RollingRecordManager{
		Record: NewRollingRecord(sublogger, bc, startSlot, &beaconCfg, rewardsInterval),

		logger:               sublogger,
		cfg:                  cfg,
		rp:                   rp,
		bc:                   bc,
		mgr:                  mgr,
		startSlot:            startSlot,
		beaconCfg:            beaconCfg,
		genesisTime:          genesisTime,
		compressor:           encoder,
		decompressor:         decoder,
		recordsFilenameRegex: recordsFilenameRegex,
	}, nil
}

// Generate a new record for the provided slot using the latest viable saved record
func (r *RollingRecordManager) GenerateRecordForState(context context.Context, state *state.NetworkState) (*RollingRecord, error) {
	// Load the latest viable record
	slot := state.BeaconSlotNumber
	rewardsInterval := state.NetworkDetails.RewardIndex
	record, err := r.LoadBestRecordFromDisk(r.startSlot, slot, rewardsInterval)
	if err != nil {
		return nil, fmt.Errorf("error loading best record for slot %d: %w", slot, err)
	}

	if record.LastDutiesSlot == slot {
		// Already have a full snapshot so we don't have to do anything
		r.logger.Info("Loaded record was already up-to-date", slog.Uint64(keys.SlotKey, slot))
		return record, nil
	} else if record.LastDutiesSlot > slot {
		// This should never happen but sanity check it anyway
		return nil, fmt.Errorf("loaded record has duties completed for slot %d, which is too far forward (targeting slot %d)", record.LastDutiesSlot, slot)
	}

	// Update to the target slot
	err = r.UpdateRecordToState(context, state, slot)
	if err != nil {
		return nil, fmt.Errorf("error updating record to slot %d: %w", slot, err)
	}

	return record, nil
}

// Save the rolling record to a file and update the record info catalog
func (r *RollingRecordManager) SaveRecordToFile(record *RollingRecord) error {

	// Serialize the record
	bytes, err := record.Serialize()
	if err != nil {
		return fmt.Errorf("error saving rolling record: %w", err)
	}

	// Compress the record
	compressedBytes := r.compressor.EncodeAll(bytes, make([]byte, 0, len(bytes)))

	// Get the record filename
	slot := record.LastDutiesSlot
	epoch := record.LastDutiesSlot / r.beaconCfg.SlotsPerEpoch
	recordsPath := r.cfg.GetRecordsPath()
	filename := filepath.Join(recordsPath, fmt.Sprintf(recordsFilenameFormat, slot, epoch))

	// Write it to a file
	err = os.WriteFile(filename, compressedBytes, 0664)
	if err != nil {
		return fmt.Errorf("error writing file [%s]: %w", filename, err)
	}

	// Compute the SHA384 hash to act as a checksum
	checksum := sha512.Sum384(compressedBytes)

	// Load the existing checksum table
	_, lines, err := r.parseChecksumFile()
	if err != nil {
		return fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if lines == nil {
		lines = []string{}
	}

	// Add the new record checksum
	baseFilename := filepath.Base(filename)
	checksumLine := fmt.Sprintf("%s  %s", hex.EncodeToString(checksum[:]), baseFilename)

	// Sort the lines by their slot
	err = r.sortChecksumEntries(lines)
	if err != nil {
		return fmt.Errorf("error sorting checkpoint file entries: %w", err)
	}

	overwritten := false
	for i, line := range lines {
		if strings.HasSuffix(line, baseFilename) {
			// If there is already a line with the filename, overwrite it
			lines[i] = checksumLine
			overwritten = true
			break
		}
	}
	if !overwritten {
		// If there's no existing lines, add this to the end
		lines = append(lines, checksumLine)
	}

	// Get the number of lines to write
	checkpointRetentionLimit := r.cfg.CheckpointRetentionLimit.Value
	var newLines []string
	if len(lines) > int(checkpointRetentionLimit) {
		numberOfNewLines := int(checkpointRetentionLimit)
		cullCount := len(lines) - numberOfNewLines

		// Remove old lines and delete the corresponding files that shouldn't be retained
		for i := 0; i < cullCount; i++ {
			line := lines[i]

			// Extract the filename
			elems := strings.Split(line, "  ")
			if len(elems) != 2 {
				return fmt.Errorf("error parsing checkpoint line (%s): expected 2 elements, but got %d", line, len(elems))
			}
			filename := elems[1]
			fullFilename := filepath.Join(recordsPath, filename)

			// Delete the file if it exists
			_, err := os.Stat(fullFilename)
			if os.IsNotExist(err) {
				r.logger.Info("NOTE: tried removing checkpoint file based on the retention limit, but it didn't exist.", slog.String(keys.FileKey, filename))
				continue
			}
			err = os.Remove(fullFilename)
			if err != nil {
				return fmt.Errorf("error deleting file [%s]: %w", fullFilename, err)
			}

			r.logger.Info("Removed checkpoint file based on the retention limit.", slog.String(keys.FileKey, filename))
		}

		// Store the rest
		newLines = make([]string, numberOfNewLines)
		for i := cullCount; i <= numberOfNewLines; i++ {
			newLines[i-cullCount] = lines[i]
		}
	} else {
		newLines = lines
	}

	fileContents := strings.Join(newLines, "\n")
	checksumBytes := []byte(fileContents)

	// Save the new file
	checksumFilename := filepath.Join(recordsPath, config.ChecksumTableFilename)
	err = os.WriteFile(checksumFilename, checksumBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing checksum file after culling: %w", err)
	}

	return nil
}

// Load the most recent appropriate rolling record from disk, using the checksum table as an index
func (r *RollingRecordManager) LoadBestRecordFromDisk(startSlot uint64, targetSlot uint64, rewardsInterval uint64) (*RollingRecord, error) {
	recordCheckpointInterval := r.cfg.RecordCheckpointInterval.Value
	latestCompatibleVersion, err := semver.New(latestCompatibleVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest compatible version string [%s]: %w", latestCompatibleVersionString, err)
	}

	// Parse the checksum file
	exists, lines, err := r.parseChecksumFile()
	if err != nil {
		return nil, fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if !exists {
		// There isn't a checksum file so start over
		r.logger.Info("Checksum file not found, creating a new record from the start of the interval.")
		record := NewRollingRecord(r.logger, r.bc, startSlot, &r.beaconCfg, rewardsInterval)
		r.Record = record
		r.nextEpochToSave = startSlot/r.beaconCfg.SlotsPerEpoch + recordCheckpointInterval - 1
		return record, nil
	}

	// Iterate over each file, counting backwards from the bottom
	recordsPath := r.cfg.GetRecordsPath()
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Extract the checksum, filename, and slot number
		checksumString, filename, slot, err := r.parseChecksumEntry(line)
		if err != nil {
			return nil, err
		}

		// Check if the slot was too far into the future
		if slot > targetSlot {
			r.logger.Info("File was too far into the future, trying an older one...", slog.String(keys.FileKey, filename))
			continue
		}

		// Check if it was too far into the past
		if slot < startSlot {
			r.logger.Warn("File was too old (generated before the target start slot), none of the remaining records can be used.", slog.String(keys.FileKey, filename))
			break
		}

		// Make sure the checksum parses properly
		checksum, err := hex.DecodeString(checksumString)
		if err != nil {
			return nil, fmt.Errorf("error scanning checkpoint line (%s): checksum (%s) could not be parsed", line, checksumString)
		}

		// Try to load it
		fullFilename := filepath.Join(recordsPath, filename)
		record, err := r.loadRecordFromFile(fullFilename, checksum)
		if err != nil {
			r.logger.Warn("Error loading record from file... attempting previous file", slog.String(keys.FileKey, fullFilename), log.Err(err))
			continue
		}

		// Check if it was for the proper interval
		if record.RewardsInterval != rewardsInterval {
			r.logger.Info("File was for the wrong rewards so it cannot be used, trying an earlier checkpoint.", slog.String(keys.FileKey, filename), slog.Uint64(keys.FileIntervalKey, record.RewardsInterval), slog.Uint64(keys.ActualIntervalKey, rewardsInterval))
			continue
		}

		// Check if it has the proper start slot
		if record.StartSlot != startSlot {
			r.logger.Info("File started on the wrong slot so it cannot be used, trying an earlier checkpoint.", slog.String(keys.FileKey, filename), slog.Uint64(keys.FileStartSlotKey, record.StartSlot), slog.Uint64(keys.ActualStartSlotKey, startSlot))
			continue
		}

		// Check if it's using a compatible version
		recordVersionString := record.SmartnodeVersion
		if recordVersionString == "" {
			recordVersionString = "1.10.0" // First release without version info
		}
		recordVersion, err := semver.New(recordVersionString)
		if err != nil {
			r.logger.Info("File had invalid version, trying an earlier checkpoint.", slog.String(keys.FileKey, filename), slog.String(keys.VersionKey, recordVersionString), log.Err(err))
			continue
		}
		if recordVersion.LT(*latestCompatibleVersion) {
			r.logger.Info("File was made with an incompatible Smart Node v so it cannot be used, trying an earlier checkpoint.", slog.String(keys.FileKey, filename), slog.String(keys.FileVersionKey, recordVersionString), slog.String(keys.HighestCompatibleKey, latestCompatibleVersionString))
			continue
		}

		epoch := slot / r.beaconCfg.SlotsPerEpoch
		r.logger.Info("Loaded latest viable rolling records file.", slog.String(keys.FileKey, filename), slog.Uint64(keys.FileStartSlotKey, slot), slog.Uint64(keys.EpochKey, epoch), slog.Uint64(keys.FileIntervalKey, record.RewardsInterval))
		r.Record = record
		r.nextEpochToSave = record.LastDutiesSlot/r.beaconCfg.SlotsPerEpoch + recordCheckpointInterval
		return record, nil

	}

	// If we got here then none of the saved files worked so we have to make a new record
	r.logger.Warn("None of the saved record checkpoint files were eligible for use, creating a new record from the start of the interval.")
	record := NewRollingRecord(r.logger, r.bc, startSlot, &r.beaconCfg, rewardsInterval)
	r.Record = record
	r.nextEpochToSave = startSlot/r.beaconCfg.SlotsPerEpoch + recordCheckpointInterval - 1
	return record, nil

}

// Updates the manager's record to the provided state, retrying upon errors until success
func (r *RollingRecordManager) UpdateRecordToState(context context.Context, state *state.NetworkState, latestFinalizedSlot uint64) error {
	err := r.updateImpl(context, state, latestFinalizedSlot)
	if err != nil {
		// Revert to the latest saved state
		r.logger.Warn("Failed to update rolling record, reverting to the last saved checkpoint to prevent corruption.", slog.Uint64(keys.BlockKey, state.ElBlockNumber), slog.Uint64(keys.SlotKey, state.BeaconSlotNumber), log.Err(err))
		_, err2 := r.LoadBestRecordFromDisk(r.startSlot, latestFinalizedSlot, r.Record.RewardsInterval)
		if err2 != nil {
			return fmt.Errorf("error loading last best checkpoint: %w", err)
		}

		// Try again
		r.logger.Info("Successfully reverted to the last saved state.")
		return err
	}

	return nil
}

// Updates the manager's record to the provided state
func (r *RollingRecordManager) updateImpl(context context.Context, state *state.NetworkState, latestFinalizedSlot uint64) error {
	var err error
	r.logger.Info("Started record update.", slog.Uint64(keys.TargetSlotKey, latestFinalizedSlot))

	// Create a new record if the current one is for the previous rewards interval
	if r.Record.RewardsInterval < state.NetworkDetails.RewardIndex {
		err := r.createNewRecord(context, state)
		if err != nil {
			return fmt.Errorf("error creating new record: %w", err)
		}
	}

	// Get the state for the target slot
	recordCheckpointInterval := r.cfg.RecordCheckpointInterval.Value
	finalTarget := latestFinalizedSlot
	finalizedState := state
	if finalTarget != state.BeaconSlotNumber {
		finalizedState, err = r.mgr.GetStateForSlot(context, finalTarget)
		if err != nil {
			return fmt.Errorf("error getting state for latest finalized slot (%d): %w", finalTarget, err)
		}
	}

	// Break the routine into chunks so it can be saved if necessary
	nextStartSlot := r.Record.LastDutiesSlot + 1
	if r.Record.LastDutiesSlot == 0 {
		nextStartSlot = r.startSlot
	}

	nextStartEpoch := nextStartSlot / r.beaconCfg.SlotsPerEpoch
	finalEpoch := finalTarget / r.beaconCfg.SlotsPerEpoch

	nextTargetEpoch := finalEpoch
	if nextTargetEpoch > r.nextEpochToSave {
		// Make a stop at the next required checkpoint so it can be saved
		nextTargetEpoch = r.nextEpochToSave
	}
	nextTargetSlot := (nextTargetEpoch+1)*r.beaconCfg.SlotsPerEpoch - 1 // Target is the last slot of the epoch
	if nextTargetSlot > finalTarget {
		nextTargetSlot = finalTarget
	}
	totalSlots := float64(finalTarget - nextStartSlot + 1)
	initialSlot := nextStartSlot

	r.logger.Info("Collecting records...", slog.Uint64(keys.StartSlotKey, nextStartSlot), slog.Uint64(keys.StartEpochKey, nextStartEpoch), slog.Uint64(keys.EndSlotKey, finalTarget), slog.Uint64(keys.EndEpochKey, finalEpoch))
	startTime := time.Now()
	for {
		if nextStartSlot > finalTarget {
			break
		}

		// Update the record to the target state
		err = r.Record.UpdateToSlot(context, nextTargetSlot, finalizedState)
		if err != nil {
			return fmt.Errorf("error updating rolling record to slot %d, block %d: %w", state.BeaconSlotNumber, state.ElBlockNumber, err)
		}
		slotsProcessed := nextTargetSlot - initialSlot + 1
		r.logger.Info(fmt.Sprintf("(%.2f%%) Updated from slot %d (epoch %d) to slot %d (epoch %d)...", float64(slotsProcessed)/totalSlots*100.0, nextStartSlot, nextStartEpoch, nextTargetSlot, nextTargetEpoch), slog.Duration(keys.TotalElapsedKey, time.Since(startTime)))

		// Save if required
		if nextTargetEpoch == r.nextEpochToSave {
			err = r.SaveRecordToFile(r.Record)
			if err != nil {
				return fmt.Errorf("error saving record: %w", err)
			}
			r.logger.Info("Saved record checkpoint.")
			r.nextEpochToSave += recordCheckpointInterval // Set the next epoch to save 1 checkpoint in the future
		}

		nextStartSlot = nextTargetSlot + 1
		nextStartEpoch = nextStartSlot / r.beaconCfg.SlotsPerEpoch
		nextTargetEpoch = finalEpoch
		if nextTargetEpoch > r.nextEpochToSave {
			// Make a stop at the next required checkpoint so it can be saved
			nextTargetEpoch = r.nextEpochToSave
		}
		nextTargetSlot = (nextTargetEpoch+1)*r.beaconCfg.SlotsPerEpoch - 1 // Target is the last slot of the epoch
		if nextTargetSlot > finalTarget {
			nextTargetSlot = finalTarget
		}
	}

	// Log the update
	startEpoch := r.Record.StartSlot / r.beaconCfg.SlotsPerEpoch
	currentEpoch := r.Record.LastDutiesSlot / r.beaconCfg.SlotsPerEpoch
	r.logger.Info("Record update complete. (slot %d-%d, epoch %d-%d).", slog.Uint64(keys.StartSlotKey, r.Record.StartSlot), slog.Uint64(keys.StartEpochKey, startEpoch), slog.Uint64(keys.EndSlotKey, r.Record.LastDutiesSlot), slog.Uint64(keys.EndEpochKey, currentEpoch))
	return nil
}

// Prepares the record for a rewards interval report
func (r *RollingRecordManager) PrepareRecordForReport(context context.Context, state *state.NetworkState) error {
	rewardsSlot := state.BeaconSlotNumber

	// Check if the current record has gone past the requested slot or if it can be updated / used
	if rewardsSlot < r.Record.LastDutiesSlot {
		r.logger.Info("Current record has extended too far, reverting to a previous checkpoint.", slog.Uint64(keys.TargetSlotKey, rewardsSlot), slog.Uint64(keys.RecordSlotKey, r.Record.LastDutiesSlot))

		newRecord, err := r.GenerateRecordForState(context, state)
		if err != nil {
			return fmt.Errorf("error creating record for rewards slot: %w", err)
		}

		r.Record = newRecord
	} else {
		r.logger.Info("Current record can be used, updating to target slot.", slog.Uint64(keys.TargetSlotKey, rewardsSlot), slog.Uint64(keys.RecordSlotKey, r.Record.LastDutiesSlot))
		err := r.UpdateRecordToState(context, state, rewardsSlot)
		if err != nil {
			return fmt.Errorf("error updating record to rewards slot: %w", err)
		}
	}

	return nil
}

// Get the slot number from a record filename
func (r *RollingRecordManager) getSlotFromFilename(filename string) (uint64, error) {
	matches := r.recordsFilenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return 0, fmt.Errorf("filename (%s) did not match the expected format", filename)
	}
	slotIndex := r.recordsFilenameRegex.SubexpIndex("slot")
	if slotIndex == -1 {
		return 0, fmt.Errorf("slot number not found in filename (%s)", filename)
	}
	slotString := matches[slotIndex]
	slot, err := strconv.ParseUint(slotString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("slot (%s) could not be parsed to a number", slotString)
	}

	return slot, nil
}

// Load a record from a file, making sure its contents match the provided checksum
func (r *RollingRecordManager) loadRecordFromFile(filename string, expectedChecksum []byte) (*RollingRecord, error) {
	// Read the file
	compressedBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Calculate the hash and validate it
	checksum := sha512.Sum384(compressedBytes)
	if !bytes.Equal(expectedChecksum, checksum[:]) {
		expectedString := hex.EncodeToString(expectedChecksum)
		actualString := hex.EncodeToString(checksum[:])
		return nil, fmt.Errorf("checksum mismatch (expected %s, but it was %s)", expectedString, actualString)
	}

	// Decompress it
	bytes, err := r.decompressor.DecodeAll(compressedBytes, []byte{})
	if err != nil {
		return nil, fmt.Errorf("error decompressing data: %w", err)
	}

	// Create a new record from the data
	return DeserializeRollingRecord(r.logger, r.bc, &r.beaconCfg, bytes)
}

// Get the lines from the checksum file
func (r *RollingRecordManager) parseChecksumFile() (bool, []string, error) {
	// Get the checksum filename
	recordsPath := r.cfg.GetRecordsPath()
	checksumFilename := filepath.Join(recordsPath, config.ChecksumTableFilename)

	// Check if the file exists
	_, err := os.Stat(checksumFilename)
	if os.IsNotExist(err) {
		return false, nil, nil
	}

	// Open the checksum file
	checksumTable, err := os.ReadFile(checksumFilename)
	if err != nil {
		return false, nil, fmt.Errorf("error loading checksum table (%s): %w", checksumFilename, err)
	}

	// Parse out each line
	originalLines := strings.Split(string(checksumTable), "\n")

	// Remove empty lines
	lines := make([]string, 0, len(originalLines))
	for _, line := range originalLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			lines = append(lines, line)
		}
	}

	return true, lines, nil
}

// Sort the checksum file entries by their slot
func (r *RollingRecordManager) sortChecksumEntries(lines []string) error {
	var sortErr error
	sort.Slice(lines, func(i int, j int) bool {
		_, _, firstSlot, err := r.parseChecksumEntry(lines[i])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		_, _, secondSlot, err := r.parseChecksumEntry(lines[j])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		return firstSlot < secondSlot
	})
	return sortErr
}

// Get the checksum, the filename, and the slot number from a checksum entry.
func (r *RollingRecordManager) parseChecksumEntry(line string) (string, string, uint64, error) {
	// Extract the checksum and filename
	elems := strings.Split(line, "  ")
	if len(elems) != 2 {
		return "", "", 0, fmt.Errorf("error parsing checkpoint line (%s): expected 2 elements, but got %d", line, len(elems))
	}
	checksumString := elems[0]
	filename := elems[1]

	// Extract the slot number for this file
	slot, err := r.getSlotFromFilename(filename)
	if err != nil {
		return "", "", 0, fmt.Errorf("error scanning checkpoint line (%s): %w", line, err)
	}

	return checksumString, filename, slot, nil
}

// Creates a new record
func (r *RollingRecordManager) createNewRecord(context context.Context, state *state.NetworkState) error {
	// Get the current interval index
	rewardsPool, err := rewards.NewRewardsPool(r.rp)
	if err != nil {
		return fmt.Errorf("error getting rewards pool binding: %w", err)
	}
	err = r.rp.Query(nil, nil, rewardsPool.RewardIndex)
	if err != nil {
		return fmt.Errorf("error getting rewards index: %w", err)
	}
	currentIndex := rewardsPool.RewardIndex.Formatted()

	// Get the last rewards event and starting epoch
	resources := r.cfg.GetRocketPoolResources()
	found, event, err := rewardsPool.GetRewardsEvent(r.rp, currentIndex-1, resources.PreviousRewardsPoolAddresses, nil)
	if err != nil {
		return fmt.Errorf("error getting event for rewards interval %d: %w", currentIndex-1, err)
	}
	if !found {
		return fmt.Errorf("event for rewards interval %d not found", currentIndex-1)
	}

	// Get the start slot of the current interval
	startSlot, err := GetStartSlotForInterval(context, event, r.bc, r.beaconCfg)
	if err != nil {
		return fmt.Errorf("error getting start slot for interval %d: %w", currentIndex, err)
	}
	newEpoch := startSlot / r.beaconCfg.SlotsPerEpoch

	// Create a new record for the start slot
	r.logger.Info("Current record is for an old interval, creating a new record.", slog.Uint64(keys.RecordIntervalKey, r.Record.RewardsInterval), slog.Uint64(keys.ActualIntervalKey, state.NetworkDetails.RewardIndex), slog.Uint64(keys.StartSlotKey, startSlot), slog.Uint64(keys.StartEpochKey, newEpoch))
	r.Record = NewRollingRecord(r.logger, r.bc, startSlot, &r.beaconCfg, state.NetworkDetails.RewardIndex)
	r.startSlot = startSlot
	recordCheckpointInterval := r.cfg.RecordCheckpointInterval.Value
	r.nextEpochToSave = startSlot/r.beaconCfg.SlotsPerEpoch + recordCheckpointInterval - 1

	return nil
}
