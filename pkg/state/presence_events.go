package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// https://discord.com/developers/docs/topics/gateway#presence-update
type PresenceUpdateEvent struct {
	*gateway.PresenceUpdateEvent
	*Base

	Old *discord.Presence
}

// undocumented
type PresencesReplaceEvent struct {
	*gateway.PresencesReplaceEvent
	*Base
}

// SessionsReplaceEvent is an undocumented user event. It's likely used for
// current user's presence updates.
type SessionsReplaceEvent struct {
	*gateway.SessionsReplaceEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#typing-start
type TypingStartEvent struct {
	*gateway.TypingStartEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#user-update
type UserUpdateEvent struct {
	*gateway.UserUpdateEvent
	*Base
}
