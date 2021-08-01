package state

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"

	"github.com/mavolin/disstate/v4/pkg/event"
)

type State struct {
	// State is the arikawa *state.State this State wraps.
	// It's Gateway field should not be used.
	*state.State
	*event.Handler

	// gateways are the *gateways.gateways managed by the State.
	// They are sorted in ascending order by their shard id.
	gateways []*gateway.Gateway
	mutex    *sync.RWMutex

	// Events is the channel all gateways send their event in.
	Events chan interface{}

	// numShards is the total number of shards.
	// This may be higher than len(gateways), if other shards are running in
	// a different process/on a different machine.
	numShards int

	rescale     func(func(*State))
	rescaleExec *uint32
}

// Options are the options used to create a State.
// All options except Token are optional.
type Options struct {
	// Token is the bot's token.
	// It mustn't be prefixed with "Bot ".
	Token string
	// Status is the status of the bot.
	//
	// Default: gateway.OnlineStatus
	Status gateway.Status
	// Activity is the activity of the bot.
	//
	// Default: nil
	Activity *discord.Activity

	// Cabinet is the store's cabinet.
	//
	// Default: defaultstore.New()
	Cabinet *store.Cabinet

	// TotalShards is the total number of shards.
	// If it is <= 0, the recommended number of shards will be used.
	TotalShards int
	// ShardIDs are the shard ids this State instance will use.
	//
	// If setting this, you also need to specify the TotalShards.
	// If ShardIDs is set, but TotalShards is not, New will panic.
	ShardIDs []int

	// Gateways are the initial gateways to use.
	// It is an alternative to TotalShards, and ShardIDs and you shouldn't set
	// both.
	Gateways []*gateway.Gateway

	// Rescale is the function called, if Discord closes any of the gateways
	// with a 4011 close code aka. 'Sharding Required'.
	//
	// Update is the function to be called after a rescaled state was
	// constructed.
	// After it, the State should be opened.
	//
	// If you don't set TotalShards, this will default to the below, unless you
	// define a custom Rescale function.
	//
	// 	func(update func(*State)) {
	//		s, err := New(o)
	//		if err != nil {
	//			return
	//		}
	//
	// 		update(s)
	//		_ = s.Open(context.Background())
	//	}
	//
	// Otherwise, you are required to set this function yourself.
	// If you don't, New will panic.
	//
	// Handlers previously added to the State will not be lost and do not need
	// to manually be readded during Rescale.
	Rescale func(update func(*State))

	// ErrorHandler is the error handler of the event handler.
	//
	// Defaults to:
	//
	//	func(err error) {
	//		log.Println("event handler: ", err.Error())
	//	}
	ErrorHandler func(error)
	// PanicHandler is the panic handler of the event handler
	//
	// Defaults to:
	//
	//	func(rec interface{}) {
	//		log.Printf("event handler: panic: %s\n", rec)
	//	}
	PanicHandler func(rec interface{})
}

func (o *Options) setDefaults() {
	if len(o.Token) == 0 {
		panic("state: Options.Token may not be empty")
	}

	o.Token = "Bot " + o.Token

	if len(o.Status) == 0 {
		o.Status = gateway.OnlineStatus
	}

	if o.Cabinet == nil {
		o.Cabinet = defaultstore.New()
	}

	if len(o.ShardIDs) > 0 && o.TotalShards <= 0 {
		panic("state: setting Options.ShardIDs requires Options.TotalShards to be set as well")
	}

	if o.TotalShards > 0 && o.Rescale == nil {
		panic("state: setting Options.TotalShards requires Options.Rescale to be set as well")
	}

	if o.TotalShards <= 0 && o.Rescale == nil && len(o.Gateways) == 0 {
		o.Rescale = func(update func(*State)) {
			s, err := New(*o)
			if err != nil {
				return
			}

			update(s)
			_ = s.Open(context.Background())
		}
	}
}

// New creates a new *State using as many gateways as recommended by Discord.
func New(o Options) (*State, error) {
	o.setDefaults()

	if len(o.Gateways) == 0 {
		botData, err := gateway.BotURL(o.Token)
		if err != nil {
			return nil, err
		}

		if o.TotalShards <= 0 {
			o.TotalShards = botData.Shards
		}

		if len(o.ShardIDs) == 0 {
			o.ShardIDs = generateShardIDs(o.TotalShards)
		}

		id := gateway.DefaultIdentifier(o.Token)
		setStartLimiters(botData, id)

		id.Presence = &gateway.UpdateStatusData{Status: o.Status}
		if o.Activity != nil {
			id.Presence.Activities = append(id.Presence.Activities, *o.Activity)
		}

		o.Gateways = make([]*gateway.Gateway, len(o.ShardIDs))
		gwURL := gateway.AddGatewayParams(botData.URL)

		for i, shardID := range o.ShardIDs {
			id.Shard = new(gateway.Shard)
			id.SetShard(shardID, o.TotalShards)
			idCp := *id

			o.Gateways[i] = gateway.NewCustomIdentifiedGateway(gwURL, &idCp)
		}
	}

	s := &State{
		State:       state.NewFromSession(session.NewWithGateway(o.Gateways[0]), o.Cabinet),
		gateways:    o.Gateways,
		mutex:       new(sync.RWMutex),
		Events:      make(chan interface{}),
		numShards:   o.Gateways[0].Identifier.Shard.NumShards(),
		rescale:     o.Rescale,
		rescaleExec: new(uint32),
	}
	s.Handler = event.NewHandler(reflect.ValueOf(s))
	s.ErrorHandler = o.ErrorHandler
	s.PanicHandler = o.PanicHandler

	for _, g := range s.gateways {
		g.Events = s.Events
		g.OnShardingRequired(s.onShardingRequired)
	}

	return s, nil
}

// FromShardID returns the *gateway.Gateway with the given shard id, or nil if
// the shard manager has no gateways with the given id.
func (s *State) FromShardID(shardID int) *gateway.Gateway {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// fast-path, also prevent nil pointer dereference if this manager manages
	// a user account
	if s.numShards == 1 {
		return s.gateways[0]
	}

	i := sort.Search(len(s.gateways), func(i int) bool {
		return s.gateways[i].Identifier.Shard.ShardID() >= shardID
	})

	if i < len(s.gateways) && s.gateways[i].Identifier.Shard.ShardID() == shardID {
		return s.gateways[i]
	}

	return nil
}

// FromGuildID returns the *gateway.Gateway managing the guild with the passed
// ID, or nil if this Manager does not manage this guild.
func (s *State) FromGuildID(guildID discord.GuildID) *gateway.Gateway {
	return s.FromShardID(int(uint64(guildID>>22) % uint64(s.numShards)))
}

// Apply applies the given function to all gateways handled by this Manager.
func (s *State) Apply(f func(g *gateway.Gateway)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, g := range s.gateways {
		f(g)
	}
}

// ApplyError is the same as Apply, but the iterator function returns an error.
// If such an error occurs, the error will be returned wrapped in an *ShardError.
//
// If all is set to true, ApplyError will apply the passed function to all
// gateways, regardless of whether an error occurs.
// If a single error occurs, it will be returned as an *ShardError, if multiple
// errors occur then they will be returned as *MultiError.
func (s *State) ApplyError(f func(g *gateway.Gateway) error, all bool) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var errs MultiError

	for _, g := range s.gateways {
		if err := f(g); err != nil {
			wrapperErr := &ShardError{
				ShardID: shardID(g),
				Source:  err,
			}

			if !all {
				return wrapperErr
			}

			errs = append(errs, wrapperErr)
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errs
	}
}

// Gateways returns the gateways managed by this Manager.
func (s *State) Gateways() []*gateway.Gateway {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cp := make([]*gateway.Gateway, len(s.gateways))
	copy(cp, s.gateways)

	return cp
}

// AddIntents adds the passed gateway.Intents to all gateways managed by the
// Manager.
func (s *State) AddIntents(i gateway.Intents) {
	s.Apply(func(g *gateway.Gateway) { g.AddIntents(i) })
}

// Open opens all gateways handled by this Manager.
// If an error occurs, Open will attempt to close all previously opened
// gateways before returning.
func (s *State) Open(ctx context.Context) error {
	s.Handler.Open(s.Events)

	err := s.ApplyError(func(g *gateway.Gateway) error { return g.Open(ctx) }, false)
	if err == nil {
		return nil
	}

	var errs MultiError
	errs = append(errs, err)

	for shardID := 0; shardID < err.(*ShardError).ShardID; shardID++ {
		if shard := s.FromShardID(shardID); shard != nil { // exists?
			if err := shard.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return errs
}

// Close closes all gateways handled by this Manager.
//
// If an error occurs, Close will attempt to close all remaining gateways
// first, before returning. If multiple errors occur during that process, a
// MultiError will be returned.
func (s *State) Close() error {
	s.Handler.Close()
	return s.ApplyError(func(g *gateway.Gateway) error { return g.Close() }, true)
}

// Pause pauses all gateways managed by this Manager.
//
// If an error occurs, Pause will attempt to pause all remaining gateways
// first, before returning. If multiple errors occur during that process, a
// MultiError will be returned.
func (s *State) Pause() error {
	return s.ApplyError(func(g *gateway.Gateway) error { return g.Pause() }, true)
}

// UpdateStatus updates the status of all gateways handled by this Manager.
//
// If an error occurs, UpdateStatus will attempt to update the status of all
// remaining gateways first, before returning. If multiple errors occur during
// that process, a MultiError will be returned.
func (s *State) UpdateStatus(d gateway.UpdateStatusData) error {
	return s.ApplyError(func(g *gateway.Gateway) error { return g.UpdateStatus(d) }, true)
}

// RequestGuildMembers is used to request all members for a guild or a list of
// guilds. When initially connecting, if you don't have the GUILD_PRESENCES
// Gateway Intent, or if the guild is over 75k members, it will only send
// members who are in voice, plus the member for you (the connecting user).
// Otherwise, if a guild has over large_threshold members (value in the Gateway
// Identify), it will only send members who are online, have a role, have a
// nickname, or are in a voice channel, and if it has under large_threshold
// members, it will send all members. If a client wishes to receive additional
// members, they need to explicitly request them via this operation. The server
// will send Guild Members Chunk events in response with up to 1000 members per
// chunk until all members that match the request have been sent.
//
// Due to privacy and infrastructural concerns with this feature, there are
// some limitations that apply:
//
// 	1. GUILD_PRESENCES intent is required to set presences = true. Otherwise,
// 	   it will always be false
// 	2. GUILD_MEMBERS intent is required to request the entire member
// 	   list — (query=‘’, limit=0<=n)
// 	3. You will be limited to requesting 1 guild_id per request
// 	4. Requesting a prefix (query parameter) will return a maximum of 100
// 	   members
//
// Requesting user_ids will continue to be limited to returning 100 members
func (s *State) RequestGuildMembers(d gateway.RequestGuildMembersData) error {
	return s.FromGuildID(d.GuildIDs[0]).RequestGuildMembers(d)
}

// onShardingRequired is the function stored as Gateway.OnShardingRequired
// in every of the Manager's gateways.
func (s *State) onShardingRequired() {
	if atomic.CompareAndSwapUint32(s.rescaleExec, 0, 1) {
		_ = s.Close()

		if s.rescale == nil {
			return
		}

		update := func(newState *State) {
			s.State = newState.State
			s.gateways = newState.gateways
			s.numShards = newState.numShards
			s.rescale = newState.rescale
		}

		s.rescale(update)
	}
}

func shardID(g *gateway.Gateway) int {
	if shard := g.Identifier.Shard; shard != nil {
		return shard.ShardID()
	}

	return 0
}

func setStartLimiters(botData *api.BotData, id *gateway.Identifier) {
	resetAt := time.Now().Add(botData.StartLimit.ResetAfter.Duration())

	// Update the burst to be the current given time and reset it back to
	// the default when the given time is reached.
	id.IdentifyGlobalLimit.SetBurst(botData.StartLimit.Remaining)
	id.IdentifyGlobalLimit.SetBurstAt(resetAt, botData.StartLimit.Total)

	// Update the maximum number of identify requests allowed per 5s.
	id.IdentifyShortLimit.SetBurst(botData.StartLimit.MaxConcurrency)
}
