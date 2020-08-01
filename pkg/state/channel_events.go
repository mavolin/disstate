package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// ================================ Channel Create ================================

// https://discord.com/developers/docs/topics/gateway#channel-create
type ChannelCreateEvent struct {
	*gateway.ChannelCreateEvent
	*Base
}

type channelCreateEventHandler func(s *State, e *ChannelCreateEvent) error

func (h channelCreateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ChannelCreateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Channel Update ================================

// https://discord.com/developers/docs/topics/gateway#channel-update
type ChannelUpdateEvent struct {
	*gateway.ChannelUpdateEvent
	*Base

	Old *discord.Channel
}

type channelUpdateEventHandler func(s *State, e *ChannelUpdateEvent) error

func (h channelUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ChannelUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Channel Delete ================================

// https://discord.com/developers/docs/topics/gateway#channel-delete
type ChannelDeleteEvent struct {
	*gateway.ChannelDeleteEvent
	*Base

	Old *discord.Channel
}

type channelDeleteEventHandler func(s *State, e *ChannelDeleteEvent) error

func (h channelDeleteEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ChannelDeleteEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Channel Pins ================================

// https://discord.com/developers/docs/topics/gateway#channel-pins-update
type ChannelPinsUpdateEvent struct {
	*gateway.ChannelPinsUpdateEvent
	*Base
}

type channelPinsUpdateEventHandler func(s *State, e *ChannelPinsUpdateEvent) error

func (h channelPinsUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ChannelPinsUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Channel Unread Update ================================

type ChannelUnreadUpdateEvent struct {
	*gateway.ChannelUnreadUpdateEvent
	*Base
}

type channelUnreadUpdateEventHandler func(s *State, e *ChannelUnreadUpdateEvent) error

func (h channelUnreadUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*ChannelUnreadUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}
