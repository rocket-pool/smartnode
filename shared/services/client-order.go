package services

import (
	"fmt"
	"strings"
)

// clientRole identifies which of the two configured clients a manager should
// try next. The order is selected by preferFallback.
type clientRole int

const (
	primaryClient clientRole = iota
	fallbackClient
)

func clientOrder(preferFallback bool) []clientRole {
	if preferFallback {
		return []clientRole{fallbackClient, primaryClient}
	}
	return []clientRole{primaryClient, fallbackClient}
}

func clientName(role clientRole) string {
	switch role {
	case primaryClient:
		return "Primary"
	case fallbackClient:
		return "Fallback"
	default:
		return "Unknown"
	}
}

func isClientReady(role clientRole, primaryReady, fallbackReady bool) bool {
	switch role {
	case primaryClient:
		return primaryReady
	case fallbackClient:
		return fallbackReady
	default:
		return false
	}
}

func setClientReady(role clientRole, primaryReady, fallbackReady *bool, ready bool) {
	switch role {
	case primaryClient:
		*primaryReady = ready
	case fallbackClient:
		*fallbackReady = ready
	}
}

func nextReadyClientName(remaining []clientRole, primaryReady, fallbackReady bool) (string, bool) {
	for _, role := range remaining {
		if isClientReady(role, primaryReady, fallbackReady) {
			return strings.ToLower(clientName(role)), true
		}
	}
	return "", false
}

// tryClients walks the primary and fallback clients in the configured order.
// call is invoked for each ready client. A disconnect error marks that client
// not-ready and continues to the next; any other error is returned immediately.
// kind is used in the "no clients were ready" message (e.g. "Beacon", "Execution").
func tryClients(
	preferFallback bool,
	primaryReady *bool,
	fallbackReady *bool,
	isDisconnected func(error) bool,
	onDisconnect func(failedName, nextName string, hasNext bool, err error),
	kind string,
	call func(role clientRole) error,
) error {
	order := clientOrder(preferFallback)
	for i, role := range order {
		if !isClientReady(role, *primaryReady, *fallbackReady) {
			continue
		}
		err := call(role)
		if err == nil {
			return nil
		}
		if !isDisconnected(err) {
			return err
		}
		setClientReady(role, primaryReady, fallbackReady, false)
		nextName, hasNext := nextReadyClientName(order[i+1:], *primaryReady, *fallbackReady)
		if onDisconnect != nil {
			onDisconnect(clientName(role), nextName, hasNext, err)
		}
	}
	return fmt.Errorf("no %s clients were ready", kind)
}
