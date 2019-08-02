package messaging

import (
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/messaging"
)


// Event data
type EventData struct {
    d string
}


// Test publisher functionality
func TestPublisher(t *testing.T) {

    // Initialise publisher
    publisher := messaging.NewPublisher()

    // Initialise listener channels
    listener1 := make(chan interface{})
    listener2 := make(chan interface{})

    // Handle listener events
    eventsReceived := make(chan struct{})
    listener1Events := []EventData{}
    listener2Events := []EventData{}
    go (func() {
        for e := range listener1 {
            listener1Events = append(listener1Events, (e).(EventData))
            eventsReceived <- struct{}{}
        }
    })()
    go (func() {
        for e := range listener2 {
            listener2Events = append(listener2Events, (e).(EventData))
            eventsReceived <- struct{}{}
        }
    })()

    // Notify of event with no listeners subscribed
    publisher.Notify("event1", EventData{"a"})
    publisher.Notify("event2", EventData{"b"})

    // Subscribe listeners
    publisher.AddSubscriber("event2", listener1)
    publisher.AddSubscriber("event2", listener2)

    // Notify of event with both listeners subscribed
    publisher.Notify("event1", EventData{"c"})
    publisher.Notify("event2", EventData{"d"})
    <-eventsReceived
    <-eventsReceived

    // Unsubscribe listener
    publisher.RemoveSubscriber("event2", listener2)

    // Notify of event with listener 1 subscribed
    publisher.Notify("event1", EventData{"e"})
    publisher.Notify("event2", EventData{"f"})
    <-eventsReceived
    publisher.Notify("event2", EventData{"g"})
    <-eventsReceived
    publisher.Notify("event2", EventData{"h"})
    <-eventsReceived

    // Check observed events
    if len(listener1Events) != 4 { t.Fatalf("Incorrect listener 1 event count: expected %d, got %d", 4, len(listener1Events)) }
    if len(listener2Events) != 1 { t.Fatalf("Incorrect listener 2 event count: expected %d, got %d", 1, len(listener2Events)) }
    if listener1Events[0].d != "d" { t.Errorf("Incorrect listener 1 event 1 value: expected %s, got %s", "d", listener1Events[0].d) }
    if listener1Events[1].d != "f" { t.Errorf("Incorrect listener 1 event 2 value: expected %s, got %s", "f", listener1Events[1].d) }
    if listener1Events[2].d != "g" { t.Errorf("Incorrect listener 1 event 3 value: expected %s, got %s", "g", listener1Events[2].d) }
    if listener1Events[3].d != "h" { t.Errorf("Incorrect listener 1 event 4 value: expected %s, got %s", "h", listener1Events[3].d) }
    if listener2Events[0].d != "d" { t.Errorf("Incorrect listener 2 event 1 value: expected %s, got %s", "d", listener2Events[0].d) }

}

