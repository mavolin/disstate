package state

type allHandler func(s *State, e interface{}) error

func (h allHandler) handle(s *State, e interface{}) error {
	return h(s, e)
}

type baseHandler func(s *State, b *Base) error

func (h baseHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*Base); ok {
		return h(s, e)
	}

	return nil
}
