package messaging

import (
    "sync"
)


// Publisher
type Publisher struct {
    subscribers map[string][]chan<- interface{}
    lock sync.Mutex
}


/**
 * Create new publisher
 */
func NewPublisher() *Publisher {
    return &Publisher{
        subscribers: make(map[string][]chan<- interface{}),
    }
}


/**
 * Add subscriber
 */
func (publisher *Publisher) AddSubscriber(event string, listener chan<- interface{}) {

    // Lock for map modification
    publisher.lock.Lock()
    defer publisher.lock.Unlock()

    // Create event entry if not set
    if _, ok := publisher.subscribers[event]; !ok {
        publisher.subscribers[event] = make([]chan<- interface{}, 0)
    }

    // Append listener
    publisher.subscribers[event] = append(publisher.subscribers[event], listener)

}


/**
 * Remove subscriber
 */
func (publisher *Publisher) RemoveSubscriber(event string, listener chan<- interface{}) {

    // Lock for map modification
    publisher.lock.Lock()
    defer publisher.lock.Unlock()

    // Cancel if event entry not set
    if _, ok := publisher.subscribers[event]; !ok {
        return
    }

    // Find and remove listener
    for i := range publisher.subscribers[event] {
        if publisher.subscribers[event][i] == listener {
            publisher.subscribers[event] = append(publisher.subscribers[event][:i], publisher.subscribers[event][i+1:]...)
            break
        }
    }

}


/**
 * Notify of event
 */
func (publisher *Publisher) Notify(event string, data interface{}) {

    // Lock during notification
    publisher.lock.Lock()
    defer publisher.lock.Unlock()

    // Cancel if event entry not set
    if _, ok := publisher.subscribers[event]; !ok {
        return
    }

    // Send to listeners without blocking
    for _, listener := range publisher.subscribers[event] {
        go (func(listener chan<- interface{}) {
            listener <- data
        })(listener)
    }

}

