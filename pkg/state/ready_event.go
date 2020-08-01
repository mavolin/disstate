package state

import "github.com/diamondburned/arikawa/gateway"

// https://discord.com/developers/docs/topics/gateway#ready
type ReadyEvent struct {
	*gateway.ReadyEvent
	*Base
}

func (e *ReadyEvent) getType() eventType { return eventTypeReady }

type readyEventHandler func(s *State, e *ReadyEvent) error

func (h readyEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ReadyEvent); ok {
		return h(s, e)
	}

	return nil
}
