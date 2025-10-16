package project

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type EventScope int

const (
	ScopeRun EventScope = iota
	ScopeContainer
	ScopeFile
)

type EventState int

const (
	StateStart EventState = iota
	StateAdvance
	StateComplete
	StateError
)

// Event is the message sent through the system.
type Event struct {
	RunID     uuid.UUID
	Scope     EventScope
	Step      string
	UnitID    string
	State     EventState
	Value     int
	Total     int
	Msg       string
	Err       error
	Timestamp time.Time
}

type EventBus struct {
	in   chan Event
	mu   sync.Mutex
	subs []chan Event
}

func NewEventBus(buffer int) *EventBus {
	b := &EventBus{in: make(chan Event, buffer)}
	go func() {
		for e := range b.in {
			b.mu.Lock()
			for _, s := range b.subs {
				select {
				case s <- e:
				default: /* drop if subscriber is slow */
				}
			}
			b.mu.Unlock()
		}
	}()
	return b
}

func (b *EventBus) Sink() chan<- Event { return b.in }

func (b *EventBus) Subscribe(buffer int) <-chan Event {
	ch := make(chan Event, buffer)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Close() { close(b.in) }
