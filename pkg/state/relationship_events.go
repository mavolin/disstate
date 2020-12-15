package state

import "github.com/diamondburned/arikawa/v2/gateway"

// undocumented
type RelationshipAddEvent struct {
	*gateway.RelationshipAddEvent
	*Base
}

// undocumented
type RelationshipRemoveEvent struct {
	*gateway.RelationshipRemoveEvent
	*Base
}
