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

// https://discord.com/developers/docs/topics/gateway#message-update
type MessageUpdateEvent struct {
	*gateway.MessageUpdateEvent
	*Base

	Old *discord.Message
}

// https://discord.com/developers/docs/topics/gateway#message-delete
type MessageDeleteEvent struct {
	*gateway.MessageDeleteEvent
	*Base

	Old *discord.Message
}

// https://discord.com/developers/docs/topics/gateway#message-delete-bulk
type MessageDeleteBulkEvent struct {
	*gateway.MessageDeleteBulkEvent
	*Base

	Old []discord.Message
}

// https://discord.com/developers/docs/topics/gateway#message-reaction-add
type MessageReactionAddEvent struct {
	*gateway.MessageReactionAddEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove
type MessageReactionRemoveEvent struct {
	*gateway.MessageReactionRemoveEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove-all
type MessageReactionRemoveAllEvent struct {
	*gateway.MessageReactionRemoveAllEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#message-reaction-remove-emoji
type MessageReactionRemoveEmojiEvent struct {
	*gateway.MessageReactionRemoveEmoji
	*Base
}

// https://discord.com/developers/docs/topics/gateway#message-ack
type MessageAckEvent struct {
	*gateway.MessageAckEvent
	*Base
}
