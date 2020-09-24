package state

import "github.com/diamondburned/arikawa/gateway"

// https://discord.com/developers/docs/topics/gateway#voice-state-update
type VoiceStateUpdateEvent struct {
	*gateway.VoiceStateUpdateEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#voice-server-update
type VoiceServerUpdateEvent struct {
	*gateway.VoiceServerUpdateEvent
	*Base
}
