package rpip

import (
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Subscribe to alerts
func SubscribeToAlerts(email string) error {

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return err
    }
    defer store.Close()

    // Store email address
    return store.Put("rpip.alerts.subscription", email)

}


// Get subscribed address
func GetAlertsSubscription() (string, error) {

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return "", err
    }
    defer store.Close()

    // Get email address
    var email string = ""
    store.Get("rpip.alerts.subscription", &email)
    return email, nil

}


// Unsubscribe from alerts
func UnsubscribeFromAlerts() error {

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return err
    }
    defer store.Close()

    // Delete email address
    store.Delete("rpip.alerts.subscription")
    return nil

}

