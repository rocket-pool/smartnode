package messaging

import (
    "testing"
)


// Event data
type EventData struct {
    a string
    b string
}


// Test publisher functionality
func TestPublisher(t *testing.T) {

    // Initialise publisher
    publisher := NewPublisher()

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
    publisher.Notify("event1", EventData{"a", "b"})
    publisher.Notify("event2", EventData{"c", "d"})

    // Subscribe listeners
    publisher.AddSubscriber("event2", listener1)
    publisher.AddSubscriber("event2", listener2)

    // Notify of event with both listeners subscribed
    publisher.Notify("event1", EventData{"e", "f"})
    publisher.Notify("event2", EventData{"g", "h"})

    // Unsubscribe listener
    publisher.RemoveSubscriber("event2", listener2)

    // Notify of event with listener 1 subscribed
    publisher.Notify("event1", EventData{"i", "j"})
    publisher.Notify("event2", EventData{"k", "l"})

    // Wait for listener threads
    for received := 0; received < 3; {
        select {
            case <-eventsReceived:
                received++
        }
    }

    // Check observed events
    if len(listener1Events) != 2 { t.Fatalf("Incorrect listener 1 event count: expected %d, got %d", 2, len(listener1Events)) }
    if len(listener2Events) != 1 { t.Fatalf("Incorrect listener 2 event count: expected %d, got %d", 1, len(listener2Events)) }
    if listener1Events[0].a != "g" { t.Errorf("Incorrect listener 1 event 1 value: expected %s, got %s", "g", listener1Events[0].a) }
    if listener1Events[1].a != "k" { t.Errorf("Incorrect listener 1 event 2 value: expected %s, got %s", "k", listener1Events[1].a) }
    if listener2Events[0].a != "g" { t.Errorf("Incorrect listener 2 event 1 value: expected %s, got %s", "g", listener2Events[0].a) }

}

