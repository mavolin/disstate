package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// ================================ Message Create ================================

// https://discord.com/developers/docs/topics/gateway#message-create
type MessageCreateEvent struct {
	*gateway.MessageCreateEvent
	*Base
}

func (e *MessageCreateEvent) getType() eventType { return eventTypeMessageCreate }

type messageCreateEventHandler func(s *State, e *MessageCreateEvent) error

func (h messageCreateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageCreateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Update ================================

// https://discord.com/developers/docs/topics/gateway#message-update
type MessageUpdateEvent struct {
	*gateway.MessageUpdateEvent
	*Base

	Old *discord.Message
}

func (e *MessageUpdateEvent) getType() eventType { return eventTypeMessageUpdate }

type messageUpdateEventHandler func(s *State, e *MessageUpdateEvent) error

func (h messageUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Delete ================================

// https://discord.com/developers/docs/topics/gateway#message-delete
type MessageDeleteEvent struct {
	*gateway.MessageDeleteEvent
	*Base

	Old *discord.Message
}

func (e *MessageDeleteEvent) getType() eventType { return eventTypeMessageDelete }

type messageDeleteEventHandler func(s *State, e *MessageDeleteEvent) error

func (h messageDeleteEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageDeleteEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Delete Bulk ================================

// https://discord.com/developers/docs/topics/gateway#message-delete-bulk
type MessageDeleteBulkEvent struct {
	*gateway.MessageDeleteBulkEvent
	*Base

	Old []discord.Message
}

func (e *MessageDeleteBulkEvent) getType() eventType { return eventTypeMessageDeleteBulk }

type messageDeleteBulkEventHandler func(s *State, e *MessageDeleteBulkEvent) error

func (h messageDeleteBulkEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageDeleteBulkEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Reaction Add ================================

// https://discord.com/developers/docs/topics/gateway#message-reaction-add
type MessageReactionAddEvent struct {
	*gateway.MessageReactionAddEvent
	*Base
}

func (e *MessageReactionAddEvent) getType() eventType { return eventTypeMessageReactionAdd }

type messageReactionAddEventHandler func(s *State, e *MessageReactionAddEvent) error

func (h messageReactionAddEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageReactionAddEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Reaction Remove ================================

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove
type MessageReactionRemoveEvent struct {
	*gateway.MessageReactionRemoveEvent
	*Base
}

func (e *MessageReactionRemoveEvent) getType() eventType { return eventTypeMessageReactionRemove }

type messageReactionRemoveEventHandler func(s *State, e *MessageReactionRemoveEvent) error

func (h messageReactionRemoveEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageReactionRemoveEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Reaction Remove All ================================

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove-all
type MessageReactionRemoveAllEvent struct {
	*gateway.MessageReactionRemoveAllEvent
	*Base
}

func (e *MessageReactionRemoveAllEvent) getType() eventType { return eventTypeMessageReactionRemoveAll }

type messageReactionRemoveAllEventHandler func(s *State, e *MessageReactionRemoveAllEvent) error

func (h messageReactionRemoveAllEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageReactionRemoveAllEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Reaction Remove Emoji ================================

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove-emoji
type MessageReactionRemoveEmojiEvent struct {
	*gateway.MessageReactionRemoveEmoji
	*Base
}

func (e *MessageReactionRemoveEmojiEvent) getType() eventType {
	return eventTypeMessageReactionRemoveEmoji
}

type messageReactionRemoveEmojiEventHandler func(s *State, e *MessageReactionRemoveEmojiEvent) error

func (h messageReactionRemoveEmojiEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageReactionRemoveEmojiEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Message Ack ================================

// https://discord.com/developers/docs/topics/gateway#message-ack
type MessageAckEvent struct {
	*gateway.MessageAckEvent
	*Base
}

func (e *MessageAckEvent) getType() eventType { return eventTypeMessageAck }

type messageAckEventHandler func(s *State, e *MessageAckEvent) error

func (h messageAckEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*MessageAckEvent); ok {
		return h(s, e)
	}

	return nil
}
