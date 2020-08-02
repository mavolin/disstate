package state

import (
	"context"
	"sync"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/session"
	"github.com/diamondburned/arikawa/state"
	"github.com/pkg/errors"

	"github.com/mavolin/disstate/internal/moreatomic"
)

type State struct {
	*session.Session
	state.Store
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

// New creates a new state.
func New(token string) (*State, error) {
	return NewWithStore(token, state.NewDefaultStore(nil))
}

// NewWithIntents creates a new state with the given gateway intents. For more
// information, refer to gateway.Intents.
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
	st = &State{
		Session:           s,
		Store:             store,
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

func (s *State) MemberDisplayName(guildID discord.GuildID, userID discord.UserID) (string, error) {
	member, err := s.Member(guildID, userID)
	if err != nil {
		return "", err
	}

	if member.Nick == "" {
		return member.User.Username, nil
	}

	return member.Nick, nil
}

func (s *State) AuthorColor(message *gateway.MessageCreateEvent) (discord.Color, error) {
	if !message.GuildID.IsValid() { // this is a dm
		return discord.DefaultMemberColor, nil
	}

	if message.Member != nil {
		guild, err := s.Guild(message.GuildID)
		if err != nil {
			return 0, err
		}
		return discord.MemberColor(*guild, *message.Member), nil
	}

	return s.MemberColor(message.GuildID, message.Author.ID)
}

func (s *State) MemberColor(guildID discord.GuildID, userID discord.UserID) (discord.Color, error) {
	var wg sync.WaitGroup

	g, gerr := s.Store.Guild(guildID)
	m, merr := s.Store.Member(guildID, userID)

	switch {
	case gerr != nil && merr != nil:
		wg.Add(1)
		go func() {
			g, gerr = s.fetchGuild(guildID)
			wg.Done()
		}()

		m, merr = s.fetchMember(guildID, userID)
	case gerr != nil:
		g, gerr = s.fetchGuild(guildID)
	case merr != nil:
		m, merr = s.fetchMember(guildID, userID)
	}

	wg.Wait()

	if gerr != nil {
		return 0, errors.Wrap(merr, "failed to get guild")
	}
	if merr != nil {
		return 0, errors.Wrap(merr, "failed to get member")
	}

	return discord.MemberColor(*g, *m), nil
}

func (s *State) Permissions(
	channelID discord.ChannelID, userID discord.UserID) (discord.Permissions, error) {

	ch, err := s.Channel(channelID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get channel")
	}

	var wg sync.WaitGroup

	g, gerr := s.Store.Guild(ch.GuildID)
	m, merr := s.Store.Member(ch.GuildID, userID)

	switch {
	case gerr != nil && merr != nil:
		wg.Add(1)
		go func() {
			g, gerr = s.fetchGuild(ch.GuildID)
			wg.Done()
		}()

		m, merr = s.fetchMember(ch.GuildID, userID)
	case gerr != nil:
		g, gerr = s.fetchGuild(ch.GuildID)
	case merr != nil:
		m, merr = s.fetchMember(ch.GuildID, userID)
	}

	wg.Wait()

	if gerr != nil {
		return 0, errors.Wrap(merr, "failed to get guild")
	}
	if merr != nil {
		return 0, errors.Wrap(merr, "failed to get member")
	}

	return discord.CalcOverwrites(*g, *ch, *m), nil
}

func (s *State) Me() (*discord.User, error) {
	u, err := s.Store.Me()
	if err == nil {
		return u, nil
	}

	u, err = s.Session.Me()
	if err != nil {
		return nil, err
	}

	return u, s.Store.MyselfSet(*u)
}

func (s *State) Channel(id discord.ChannelID) (*discord.Channel, error) {
	c, err := s.Store.Channel(id)
	if err == nil {
		return c, nil
	}

	c, err = s.Session.Channel(id)
	if err != nil {
		return nil, err
	}

	return c, s.Store.ChannelSet(*c)
}

func (s *State) Channels(guildID discord.GuildID) ([]discord.Channel, error) {
	c, err := s.Store.Channels(guildID)
	if err == nil {
		return c, nil
	}

	c, err = s.Session.Channels(guildID)
	if err != nil {
		return nil, err
	}

	for _, ch := range c {
		ch := ch

		if err := s.Store.ChannelSet(ch); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *State) CreatePrivateChannel(recipient discord.UserID) (*discord.Channel, error) {
	c, err := s.Store.CreatePrivateChannel(recipient)
	if err == nil {
		return c, nil
	}

	c, err = s.Session.CreatePrivateChannel(recipient)
	if err != nil {
		return nil, err
	}

	return c, s.Store.ChannelSet(*c)
}

func (s *State) PrivateChannels() ([]discord.Channel, error) {
	c, err := s.Store.PrivateChannels()
	if err == nil {
		return c, nil
	}

	c, err = s.Session.PrivateChannels()
	if err != nil {
		return nil, err
	}

	for _, ch := range c {
		ch := ch

		if err := s.Store.ChannelSet(ch); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *State) Emoji(
	guildID discord.GuildID, emojiID discord.EmojiID) (*discord.Emoji, error) {

	e, err := s.Store.Emoji(guildID, emojiID)
	if err == nil {
		return e, nil
	}

	es, err := s.Session.Emojis(guildID)
	if err != nil {
		return nil, err
	}

	if err := s.Store.EmojiSet(guildID, es); err != nil {
		return nil, err
	}

	for _, e := range es {
		if e.ID == emojiID {
			return &e, nil
		}
	}

	return nil, state.ErrStoreNotFound
}

func (s *State) Emojis(guildID discord.GuildID) ([]discord.Emoji, error) {
	e, err := s.Store.Emojis(guildID)
	if err == nil {
		return e, nil
	}

	es, err := s.Session.Emojis(guildID)
	if err != nil {
		return nil, err
	}

	return es, s.Store.EmojiSet(guildID, es)
}

func (s *State) Guild(id discord.GuildID) (*discord.Guild, error) {
	c, err := s.Store.Guild(id)
	if err == nil {
		return c, nil
	}

	return s.fetchGuild(id)
}

// Guilds will only fill a maximum of 100 guilds from the API.
func (s *State) Guilds() ([]discord.Guild, error) {
	c, err := s.Store.Guilds()
	if err == nil {
		return c, nil
	}

	c, err = s.Session.Guilds(state.MaxFetchGuilds)
	if err != nil {
		return nil, err
	}

	for _, ch := range c {
		ch := ch

		if err := s.Store.GuildSet(ch); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *State) Member(guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	m, err := s.Store.Member(guildID, userID)
	if err == nil {
		return m, nil
	}

	return s.fetchMember(guildID, userID)
}

func (s *State) Members(guildID discord.GuildID) ([]discord.Member, error) {
	ms, err := s.Store.Members(guildID)
	if err == nil {
		return ms, nil
	}

	ms, err = s.Session.Members(guildID, state.MaxFetchMembers)
	if err != nil {
		return nil, err
	}

	for _, m := range ms {
		if err := s.Store.MemberSet(guildID, m); err != nil {
			return nil, err
		}
	}

	return ms, s.Gateway.RequestGuildMembers(gateway.RequestGuildMembersData{
		GuildID:   []discord.GuildID{guildID},
		Presences: true,
	})
}

func (s *State) Message(
	channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {

	m, err := s.Store.Message(channelID, messageID)
	if err == nil {
		return m, nil
	}

	var wg sync.WaitGroup

	c, cerr := s.Store.Channel(channelID)
	if cerr != nil {
		wg.Add(1)
		go func() {
			c, cerr = s.Session.Channel(channelID)
			if cerr == nil {
				cerr = s.Store.ChannelSet(*c)
			}

			wg.Done()
		}()
	}

	m, err = s.Session.Message(channelID, messageID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch message")
	}

	wg.Wait()

	if cerr != nil {
		return nil, errors.Wrap(cerr, "unable to fetch channel")
	}

	m.ChannelID = c.ID
	m.GuildID = c.GuildID

	return m, s.Store.MessageSet(*m)
}

// Messages fetches maximum 100 messages from the API, if it has to. There is no
// limit if it'Store from the State storage.
func (s *State) Messages(channelID discord.ChannelID) ([]discord.Message, error) {
	var maxMsgs = s.MaxMessages()

	ms, err := s.Store.Messages(channelID)
	if err == nil {
		// If the state already has as many messages as it can, skip the API.
		if maxMsgs <= len(ms) {
			return ms, nil
		}

		// Is the channel tiny?
		s.fewMutex.Lock()
		if _, ok := s.fewMessages[channelID]; ok {
			s.fewMutex.Unlock()
			return ms, nil
		}

		// No, fetch from the state.
		s.fewMutex.Unlock()
	}

	ms, err = s.Session.Messages(channelID, uint(maxMsgs))
	if err != nil {
		return nil, err
	}

	// New messages fetched weirdly does not have GuildID filled. We'll try and
	// get it for consistency with incoming message creates.
	var guildID discord.GuildID

	// A bit too convoluted, but whatever.
	c, err := s.Channel(channelID)
	if err == nil {
		// If it'Store 0, it'Store 0 anyway. We don't need a check here.
		guildID = c.GuildID
	}

	// Iterate in reverse, since the store is expected to prepend the latest
	// messages.
	for i := len(ms) - 1; i >= 0; i-- {
		// Set the guild ID, fine if it'Store 0 (it'Store already 0 anyway).
		ms[i].GuildID = guildID

		if err := s.Store.MessageSet(ms[i]); err != nil {
			return nil, err
		}
	}

	if len(ms) < maxMsgs {
		// Tiny channel, store this.
		s.fewMutex.Lock()
		s.fewMessages[channelID] = struct{}{}
		s.fewMutex.Unlock()

		return ms, nil
	}

	// Since the latest messages are at the end and we already know the maxMsgs,
	// we could slice this right away.
	return ms[:maxMsgs], nil
}

// Presence checks the state for user presences. If no guildID is given, it will
// look for the presence in all guilds.
func (s *State) Presence(
	guildID discord.GuildID, userID discord.UserID) (*discord.Presence, error) {

	p, err := s.Store.Presence(guildID, userID)
	if err == nil {
		return p, nil
	}

	// If there'Store no guild ID, look in all guilds
	if !guildID.IsValid() {
		g, err := s.Guilds()
		if err != nil {
			return nil, err
		}

		for _, g := range g {
			if p, err := s.Store.Presence(g.ID, userID); err == nil {
				return p, nil
			}
		}
	}

	return nil, err
}

func (s *State) Role(guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error) {
	r, err := s.Store.Role(guildID, roleID)
	if err == nil {
		return r, nil
	}

	rs, err := s.Session.Roles(guildID)
	if err != nil {
		return nil, err
	}

	var role *discord.Role

	for _, r := range rs {
		r := r

		if r.ID == roleID {
			role = &r
		}

		if err := s.RoleSet(guildID, r); err != nil {
			return role, err
		}
	}

	return role, nil
}

func (s *State) Roles(guildID discord.GuildID) ([]discord.Role, error) {
	rs, err := s.Store.Roles(guildID)
	if err == nil {
		return rs, nil
	}

	rs, err = s.Session.Roles(guildID)
	if err != nil {
		return nil, err
	}

	for _, r := range rs {
		r := r

		if err := s.RoleSet(guildID, r); err != nil {
			return rs, err
		}
	}

	return rs, nil
}

func (s *State) fetchGuild(id discord.GuildID) (g *discord.Guild, err error) {
	g, err = s.Session.Guild(id)
	if err == nil {
		err = s.Store.GuildSet(*g)
	}

	return
}

func (s *State) fetchMember(
	guildID discord.GuildID, userID discord.UserID) (m *discord.Member, err error) {

	m, err = s.Session.Member(guildID, userID)
	if err == nil {
		err = s.Store.MemberSet(guildID, *m)
	}

	return
}
