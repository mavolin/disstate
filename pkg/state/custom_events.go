package state

// ================================ Close ================================

// CloseEvent gets dispatched when the gateway closes.
type CloseEvent struct {
	*Base
}

type closeEventHandler func(s *State, e *CloseEvent) error

func (h closeEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*CloseEvent); ok {
		return h(s, e)
	}

	return nil
}