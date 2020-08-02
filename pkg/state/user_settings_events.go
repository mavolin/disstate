package state

import "github.com/diamondburned/arikawa/gateway"

// ================================ User Guild Settings Update ================================

// undocumented
type UserGuildSettingsUpdateEvent struct {
	*gateway.UserGuildSettingsUpdateEvent
	*Base
}

func (e *UserGuildSettingsUpdateEvent) getType() eventType { return eventTypeUserGuildSettingsUpdate }

type userGuildSettingsUpdateEventHandler func(s *State, e *UserGuildSettingsUpdateEvent) error

func (h userGuildSettingsUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*UserGuildSettingsUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ User Settings Update ================================

// undocumented
type UserSettingsUpdateEvent struct {
	*gateway.UserSettingsUpdateEvent
	*Base
}

func (e *UserSettingsUpdateEvent) getType() eventType { return eventTypeUserSettingsUpdate }

type userSettingsUpdateEventHandler func(s *State, e *UserSettingsUpdateEvent) error

func (h userSettingsUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*UserSettingsUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ User Note Update ================================

// undocumented
type UserNoteUpdateEvent struct {
	*gateway.UserNoteUpdateEvent
	*Base
}

func (e *UserNoteUpdateEvent) getType() eventType { return eventTypeUserNoteUpdate }

type userNoteUpdateEventHandler func(s *State, e *UserNoteUpdateEvent) error

func (h userNoteUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*UserNoteUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}
