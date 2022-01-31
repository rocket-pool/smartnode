package hex

// Add a prefix to a hex string if not present
func AddPrefix(value string) string {
	if len(value) < 2 || value[0:2] != "0x" {
		return "0x" + value
	}
	return value
}

// Remove a prefix from a hex string if present
func RemovePrefix(value string) string {
	if len(value) >= 2 && value[0:2] == "0x" {
		return value[2:]
	}
	return value
}
