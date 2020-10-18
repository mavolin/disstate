package state

import (
	"context"
	"sync"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/session"
	"github.com/diamondburned/arikawa/state"
	"github.com/pkg/errors"

	"github.com/mavolin/disstate/v2/internal/moreatomic"
)

type State struct {
	*state.State
	*EventHandler

	Ready *gateway.ReadyEvent

	StateLog func(error)

	// List of channels with few messages, so it doesn't bother hitting the API
	// again.
	fewMessages map[discord.ChannelID]struct{}
	fewMutex    *sync.Mutex

	// unavailableGuilds is a set of discord.GuildIDs of guilds that became
	// unavailable when already connected to the gateway, i.e. sent in a
	// GuildUnavailableEvent.
	unavailableGuilds *moreatomic.GuildIDSet
	// unreadyGuilds is a set of discord.GuildIDs of guilds that were
	// unavailable when connecting to the gateway, i.e. they had Unavailable
	// set to true during Ready.
	unreadyGuilds *moreatomic.GuildIDSet
}

// New creates a new State using the passed token.
// If creating a bot session, the token must start with 'Bot '.
func New(token string) (*State, error) {
	return NewWithStore(token, state.NewDefaultStore(nil))
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

	return NewFromSession(s, state.NewDefaultStore(nil)), nil
}

// NewWithStore creates a new State with a custom state.Store.
func NewWithStore(token string, store state.Store) (*State, error) {
	s, err := session.New(token)
	if err != nil {
		return nil, err
	}

	return NewFromSession(s, store), nil
}

// NewFromSession creates a new *State from the passed Session.
// The Session may not be opened.
func NewFromSession(s *session.Session, store state.Store) (st *State) {
	src, _ := state.NewFromSession(s, store) // doc guarantees no error

	st = &State{
		State:             src,
		StateLog:          func(error) {},
		fewMessages:       map[discord.ChannelID]struct{}{},
		fewMutex:          new(sync.Mutex),
		unavailableGuilds: moreatomic.NewGuildIDSet(),
		unreadyGuilds:     moreatomic.NewGuildIDSet(),
	}

	st.EventHandler = NewEventHandler(st)

	return
}

// NewFromState creates a new State based on a arikawa State.
// Event handlers from the old state won't be copied.
func NewFromState(s *state.State) (st *State) {
	st = &State{
		State:             s,
		StateLog:          func(error) {},
		fewMessages:       map[discord.ChannelID]struct{}{},
		fewMutex:          new(sync.Mutex),
		unavailableGuilds: moreatomic.NewGuildIDSet(),
		unreadyGuilds:     moreatomic.NewGuildIDSet(),
	}

	st.EventHandler = NewEventHandler(st)

	return
}

// Open opens a connection to the gateway.
func (s *State) Open() error {
	s.EventHandler.Open(s.Gateway.Events)

	if err := s.Gateway.Open(); err != nil {
		return errors.Wrap(err, "failed to start gateway")
	}

	return nil
}

// Close closes the connection to the gateway and stops listening for events.
func (s *State) Close() (err error) {
	err = s.Gateway.Close()

	s.EventHandler.Close()

	s.Call(&CloseEvent{
		Base: NewBase(),
	})

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

func (s *State) AuthorDisplayName(message *gateway.MessageCreateEvent) string {
	if !message.GuildID.IsValid() {
		return message.Author.Username
	}

	if message.Member != nil {
		if message.Member.Nick != "" {
			return message.Member.Nick
		}
		return message.Author.Username
	}

	n, err := s.MemberDisplayName(message.GuildID, message.Author.ID)
	if err != nil {
		return message.Author.Username
	}

	return n
}
