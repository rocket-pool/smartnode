package rpip

import (
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Subscribe to alerts
func subscribeToAlerts(email string) error {

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
func getAlertsSubscription() (string, error) {

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
func unsubscribeFromAlerts() error {

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

