package state

import "github.com/diamondburned/arikawa/gateway"

// ================================ Invite Create ================================

// https://discord.com/developers/docs/topics/gateway#invite-create
type InviteCreateEvent struct {
	*gateway.InviteCreateEvent
	*Base
}

func (e *InviteCreateEvent) getType() eventType { return eventTypeInviteCreate }

type inviteCreateEventHandler func(s *State, e *InviteCreateEvent) error

func (h inviteCreateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*InviteCreateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Invite Delete ================================

// https://discord.com/developers/docs/topics/gateway#invite-delete
type InviteDeleteEvent struct {
	*gateway.InviteDeleteEvent
	*Base
}

func (e *InviteDeleteEvent) getType() eventType { return eventTypeInviteDelete }

type inviteDeleteEventHandler func(s *State, e *InviteDeleteEvent) error

func (h inviteDeleteEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*InviteDeleteEvent); ok {
		return h(s, e)
	}

	return nil
}
