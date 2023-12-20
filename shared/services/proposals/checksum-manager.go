package proposals

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-json"
	"github.com/klauspost/compress/zstd"
)

type IDataType interface {
	GetFilename() string
}

type IChecksumDataHandler[ContextType any, DataType IDataType] interface {
	Less(firstFilename string, secondFilename string) (bool, error)
	ShouldLoadEntry(filename string, context ContextType) (bool, error)
	IsDataValid(data *DataType, filename string, context ContextType) (bool, error)
}

type ChecksumManager[ContextType any, DataType IDataType] struct {
	compressor       *zstd.Encoder
	decompressor     *zstd.Decoder
	checksumFilename string
	dataHandler      IChecksumDataHandler[ContextType, DataType]
}

func NewChecksumManager[ContextType any, DataType IDataType](checksumFilename string, dataHandler IChecksumDataHandler[ContextType, DataType]) (*ChecksumManager[ContextType, DataType], error) {
	// Create the zstd compressor and decompressor
	compressor, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, fmt.Errorf("error creating zstd compressor: %w", err)
	}
	decompressor, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd decompressor: %w", err)
	}

	// Make sure the checksum path exists
	path := filepath.Dir(checksumFilename)
	err = os.MkdirAll(path, 0775)
	if err != nil {
		return nil, fmt.Errorf("error creating folder for checksum table and data: %w", err)
	}

	return &ChecksumManager[ContextType, DataType]{
		compressor:       compressor,
		decompressor:     decompressor,
		checksumFilename: checksumFilename,
		dataHandler:      dataHandler,
	}, nil
}

func SaveToFile[ContextType any, DataType IDataType](m *ChecksumManager[ContextType, DataType], data *DataType) error {
	// Serialize the snapshot
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error serializing data: %w", err)
	}

	// Compress the snapshot
	compressedBytes := m.compressor.EncodeAll(bytes, make([]byte, 0, len(bytes)))

	// Get the baseFilename
	baseFilename := (*data).GetFilename()
	fullFilename := filepath.Join(filepath.Dir(m.checksumFilename), baseFilename)

	// Write it to a file
	err = os.WriteFile(fullFilename, compressedBytes, 0664)
	if err != nil {
		return fmt.Errorf("error writing file [%s]: %w", fullFilename, err)
	}

	// Compute the SHA384 hash to act as a checksum
	checksum := sha512.Sum384(compressedBytes)

	// Load the existing checksum table
	_, lines, err := parseChecksumFile(m.checksumFilename)
	if err != nil {
		return fmt.Errorf("error parsing checksum file: %w", err)
	}
	if lines == nil {
		lines = []string{}
	}

	// Add the new snapshot checksum
	checksumLine := fmt.Sprintf("%s  %s", hex.EncodeToString(checksum[:]), baseFilename)

	// Sort the lines by their slot
	err = sortChecksumEntries(lines, m.dataHandler)
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

	fileContents := strings.Join(lines, "\n")
	checksumBytes := []byte(fileContents)

	// Save the new file
	err = os.WriteFile(m.checksumFilename, checksumBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing checksum file: %w", err)
	}

	return nil
}

// Load the snapshot for the provided block from disk if it exists, using the checksum table as an index
func LoadFromFile[ContextType any, DataType IDataType](m *ChecksumManager[ContextType, DataType], context ContextType) (*DataType, string, error) {
	// Parse the checksum file
	exists, lines, err := parseChecksumFile(m.checksumFilename)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if !exists {
		// There isn't a checksum file so start over
		return nil, "", nil
	}

	// Iterate over each file, counting backwards from the bottom
	dataFolder := filepath.Dir(m.checksumFilename)
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Get the checksum from the line
		checksumString, filename, found := strings.Cut(line, "  ")
		if !found {
			return nil, "", fmt.Errorf("error parsing checkpoint line (%s): invalid format", line)
		}

		// Check if this is the entry to load
		shouldLoad, err := m.dataHandler.ShouldLoadEntry(filename, context)
		if err != nil {
			return nil, "", err
		}
		if !shouldLoad {
			continue
		}

		// Make sure the checksum parses properly
		savedChecksum, err := hex.DecodeString(checksumString)
		if err != nil {
			return nil, "", fmt.Errorf("error scanning checkpoint line (%s): checksum (%s) could not be parsed", line, checksumString)
		}

		// Read the file
		var data DataType
		fullFilename := filepath.Join(dataFolder, filename)
		compressedBytes, err := os.ReadFile(fullFilename)
		if err != nil {
			return nil, "", fmt.Errorf("error reading file: %w", err)
		}

		// Calculate the hash and validate it
		calculatedChecksum := sha512.Sum384(compressedBytes)
		if !bytes.Equal(savedChecksum, calculatedChecksum[:]) {
			actualString := hex.EncodeToString(calculatedChecksum[:])
			return nil, "", fmt.Errorf("checksum mismatch (expected %s, but it was %s)", checksumString, actualString)
		}

		// Decompress it
		bytes, err := m.decompressor.DecodeAll(compressedBytes, []byte{})
		if err != nil {
			return nil, "", fmt.Errorf("error decompressing data: %w", err)
		}

		// Deserialize it
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			return nil, "", fmt.Errorf("error deserializing data: %w", err)
		}

		// Check if it's valid; if so, we can stop here
		valid, err := m.dataHandler.IsDataValid(&data, filename, context)
		if err != nil {
			return nil, "", fmt.Errorf("error checking data for validity: %w", err)
		}
		if valid {
			return &data, filename, nil
		}
	}

	return nil, "", nil
}

// Get the lines from the checksum file
func parseChecksumFile(checksumFilename string) (bool, []string, error) {
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

func sortChecksumEntries[ContextType any, DataType IDataType](lines []string, dataHandler IChecksumDataHandler[ContextType, DataType]) error {
	var sortErr error
	sort.Slice(lines, func(i int, j int) bool {
		if sortErr != nil {
			return false
		}

		firstLine := lines[i]
		_, firstFilename, success := strings.Cut(firstLine, "  ")
		if !success {
			sortErr = fmt.Errorf("error parsing checkpoint line (%s): invalid format", firstLine)
			return false
		}

		secondLine := lines[j]
		_, secondFilename, success := strings.Cut(secondLine, "  ")
		if !success {
			sortErr = fmt.Errorf("error parsing checkpoint line (%s): invalid format", secondLine)
			return false
		}

		isLess, err := dataHandler.Less(firstFilename, secondFilename)
		if err != nil {
			sortErr = err
			return false
		}

		return isLess
	})
	return sortErr
}
