package state

import "github.com/diamondburned/arikawa/v2/gateway"

// https://discord.com/developers/docs/topics/gateway#invite-create
type InviteCreateEvent struct {
	*gateway.InviteCreateEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#invite-delete
type InviteDeleteEvent struct {
	*gateway.InviteDeleteEvent
	*Base
}
