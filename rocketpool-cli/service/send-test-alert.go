package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func sendTestAlert() error {
	rp := rocketpool.NewClient()
	defer rp.Close()

	_, err := rp.NodeSendTestAlert()
	if err != nil {
		return fmt.Errorf("error sending test alert: %w", err)
	}

	fmt.Println("Test alert sent successfully. Check your configured notification channels (Discord, email, Pushover) to confirm delivery.")
	return nil
}
