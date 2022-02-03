package net

import (
	"fmt"
	"regexp"
)

// Add a default port to a host address
func DefaultPort(host string, port string) string {
	if !regexp.MustCompile(":\\d+$").MatchString(host) {
		return fmt.Sprintf("%s:%s", host, port)
	}
	return host
}
