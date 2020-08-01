package state

import "github.com/diamondburned/arikawa/gateway"

// ================================ Relationship Add ================================

// undocumented
type RelationshipAddEvent struct {
	*gateway.RelationshipAddEvent
	*Base
}

func (e *RelationshipAddEvent) getType() eventType { return eventTypeRelationshipAdd }

type relationshipAddEventHandler func(s *State, e *RelationshipAddEvent) error

func (h relationshipAddEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*RelationshipAddEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Relationship Remove ================================

// undocumented
type RelationshipRemoveEvent struct {
	*gateway.RelationshipRemoveEvent
	*Base
}

func (e *RelationshipRemoveEvent) getType() eventType { return eventTypeRelationshipRemove }

type relationshipRemoveEventHandler func(s *State, e *RelationshipRemoveEvent) error

func (h relationshipRemoveEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*RelationshipRemoveEvent); ok {
		return h(s, e)
	}

	return nil
}
