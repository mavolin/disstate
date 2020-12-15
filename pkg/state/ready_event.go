package state

import "github.com/diamondburned/arikawa/v2/gateway"

// https://discord.com/developers/docs/topics/gateway#ready
type ReadyEvent struct {
	*gateway.ReadyEvent
	*Base
}
