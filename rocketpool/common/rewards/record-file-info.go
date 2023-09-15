package rewards

// Information about a saved rolling record
type RecordFileInfo struct {
	StartSlot uint64   `json:"startSlot"`
	EndSlot   uint64   `json:"endSlot"`
	Filename  string   `json:"filename"`
	Version   int      `json:"version"`
	Checksum  [48]byte `json:"checksum"`
}
