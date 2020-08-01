package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// ================================ Presence Update ================================

// https://discord.com/developers/docs/topics/gateway#presence-update
type PresenceUpdateEvent struct {
	*gateway.PresenceUpdateEvent
	*Base

	Old *discord.Presence
}

func (e *PresenceUpdateEvent) getType() eventType { return eventTypePresenceUpdate }

type presenceUpdateEventHandler func(s *State, e *PresenceUpdateEvent) error

func (h presenceUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*PresenceUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Presences Replace ================================

// undocumented
type PresencesReplaceEvent struct {
	*gateway.PresencesReplaceEvent
	*Base
}

func (e *PresencesReplaceEvent) getType() eventType { return eventTypePresencesReplace }

type presencesReplaceEventHandler func(s *State, e *PresencesReplaceEvent) error

func (h presencesReplaceEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*PresencesReplaceEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Sessions Replace ================================

// SessionsReplaceEvent is an undocumented user event. It's likely used for
// current user's presence updates.
type SessionsReplaceEvent struct {
	*gateway.SessionsReplaceEvent
	*Base
}

func (e *SessionsReplaceEvent) getType() eventType { return eventTypeSessionsReplace }

type sessionsReplaceEventHandler func(s *State, e *SessionsReplaceEvent) error

func (h sessionsReplaceEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*SessionsReplaceEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Typing Start ================================

// https://discord.com/developers/docs/topics/gateway#typing-start
type TypingStartEvent struct {
	*gateway.TypingStartEvent
	*Base
}

func (e *TypingStartEvent) getType() eventType { return eventTypeTypingStart }

type typingStartEventHandler func(s *State, e *TypingStartEvent) error

func (h typingStartEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*TypingStartEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ User Update ================================

// https://discord.com/developers/docs/topics/gateway#user-update
type UserUpdateEvent struct {
	*gateway.UserUpdateEvent
	*Base
}

func (e *UserUpdateEvent) getType() eventType { return eventTypeUserUpdate }

type userUpdateEventHandler func(s *State, e *UserUpdateEvent) error

func (h userUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*UserUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}
