package state

import "github.com/diamondburned/arikawa/gateway"

// ================================ Voice State Update ================================

// https://discord.com/developers/docs/topics/gateway#voice-state-update
type VoiceStateUpdateEvent struct {
	*gateway.VoiceStateUpdateEvent
	*Base
}

func (e *VoiceStateUpdateEvent) getType() eventType { return eventTypeVoiceStateUpdate }

type voiceStateUpdateEventHandler func(s *State, e *VoiceStateUpdateEvent) error

func (h voiceStateUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*VoiceStateUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Voice Server Update ================================

// https://discord.com/developers/docs/topics/gateway#voice-server-update
type VoiceServerUpdateEvent struct {
	*gateway.VoiceServerUpdateEvent
	*Base
}

func (e *VoiceServerUpdateEvent) getType() eventType { return eventTypeVoiceServerUpdate }

type voiceServerUpdateEventHandler func(s *State, e *VoiceServerUpdateEvent) error

func (h voiceServerUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*VoiceServerUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}
