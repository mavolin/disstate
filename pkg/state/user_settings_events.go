package state

import "github.com/diamondburned/arikawa/v2/gateway"

// undocumented
type UserGuildSettingsUpdateEvent struct {
	*gateway.UserGuildSettingsUpdateEvent
	*Base
}

// undocumented
type UserSettingsUpdateEvent struct {
	*gateway.UserSettingsUpdateEvent
	*Base
}

// undocumented
type UserNoteUpdateEvent struct {
	*gateway.UserNoteUpdateEvent
	*Base
}
