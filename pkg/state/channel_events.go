package state

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

// https://discord.com/developers/docs/topics/gateway#channel-create
type ChannelCreateEvent struct {
	*gateway.ChannelCreateEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#channel-update
type ChannelUpdateEvent struct {
	*gateway.ChannelUpdateEvent
	*Base

	Old *discord.Channel
}

// https://discord.com/developers/docs/topics/gateway#channel-delete
type ChannelDeleteEvent struct {
	*gateway.ChannelDeleteEvent
	*Base

	Old *discord.Channel
}

// https://discord.com/developers/docs/topics/gateway#channel-pins-update
type ChannelPinsUpdateEvent struct {
	*gateway.ChannelPinsUpdateEvent
	*Base
}

type ChannelUnreadUpdateEvent struct {
	*gateway.ChannelUnreadUpdateEvent
	*Base
}
