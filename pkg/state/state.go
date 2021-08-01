package state

import (
	"context"
	"reflect"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/pkg/errors"

	"github.com/mavolin/disstate/v4/pkg/event"
)

type State struct {
	*state.State
	*event.Handler
}

// New creates a new State using the passed token.
// If creating a bot session, the token must start with 'Bot '.
func New(token string) (*State, error) {
	return NewWithCabinet(token, defaultstore.New())
}

// NewWithIntents creates a new State with the given gateway intents using the
// passed token.
// If creating a bot session, the token must start with 'Bot '.
// For more information, refer to gateway.Intents.
func NewWithIntents(token string, intents ...gateway.Intents) (*State, error) {
	s, err := session.NewWithIntents(token, intents...)
	if err != nil {
		return nil, err
	}

	return NewFromSession(s, defaultstore.New()), nil
}

// NewWithCabinet creates a new State with a custom state.Store.
func NewWithCabinet(token string, cabinet *store.Cabinet) (*State, error) {
	s, err := session.New(token)
	if err != nil {
		return nil, err
	}

	return NewFromSession(s, cabinet), nil
}

// NewFromSession creates a new *State from the passed Session.
// The Session may not be opened.
func NewFromSession(s *session.Session, cabinet *store.Cabinet) (st *State) {
	st = &State{State: state.NewFromSession(s, cabinet)}

	st.Handler = event.NewHandler(reflect.ValueOf(st))

	return
}

// NewFromState creates a new State based on a arikawa State.
// Event handlers from the old state won't be copied.
func NewFromState(s *state.State) (st *State) {
	st = &State{State: s}

	st.Handler = event.NewHandler(reflect.ValueOf(st))

	return
}

// WithContext returns a shallow copy of State with the context replaced in the
// API client. All methods called on the State will use this given context. This
// method is thread-safe.
func (s *State) WithContext(ctx context.Context) *State {
	copied := *s
	copied.Client = copied.Client.WithContext(ctx)

	return &copied
}

// Open opens a connection to the gateway.
func (s *State) Open(ctx context.Context) error {
	s.Handler.Open(s.Gateway.Events)

	if err := s.Gateway.Open(ctx); err != nil {
		return errors.Wrap(err, "failed to start gateway")
	}

	return nil
}

// Close closes the connection to the gateway and stops listening for events.
func (s *State) Close() (err error) {
	err = s.Gateway.Close()

	s.Handler.Close()

	s.Call(&event.Close{Base: event.NewBase()})
	return
}

// AddIntents adds the passed intents to the state.
func (s *State) AddIntents(i gateway.Intents) {
	s.Gateway.AddIntents(i)
}
