package state

import (
	"context"
	"errors"
	"log"
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
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/diamondburned/arikawa/v3/utils/httputil"

	"github.com/mavolin/disstate/v4/pkg/event"
)

type State struct {
	// State is the arikawa *state.State this State wraps.
	// It's Gateway field should not be used.
	*state.State
	*event.Handler

	options Options

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

	rescale     func(func(Options) (*State, error))
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
	Status discord.Status
	// Activity is the activity of the bot.
	//
	// To set this to nil when calling update during rescale, set this to an
	// empty activity.
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

	// HTTPClient is the http client that will be used to make requests.
	//
	// Default: httputil.NewClient()
	HTTPClient *httputil.Client

	// Rescale is the function called, if Discord closes any of the gateways
	// with a 4011 close code aka. 'Sharding Required'.
	//
	// Usage
	//
	// To update the state's shard manager, you must call update. All
	// zero-value options in the Options you give to update, will be set to the
	// options you used when initially creating the state. However, this does
	// not apply to TotalShards, ShardIDs, and Gateways. Furthermore, setting
	// ErrorHandler or PanicHandler will have no effect.
	//
	// After calling update, you should reopen the state, by calling Open.
	// Alternatively, you can call open individually for State.Gateways().
	// Note, however, that you should call Sate.Handler.Open(State.Events),
	// before calling Gateway.Open, should you choose the individual solution.
	//
	// During update, the state's State field will be replaced, as well as the
	// gateways and the rescale function. The event handler will remain
	// untouched, which is why you don't need to readd your handlers.
	//
	// Default
	//
	// If you don't set TotalShards and Gateways, this will default to the
	// below, unless you define a custom Rescale function.
	//
	// 	func(update func(Options) *State) {
	//		s, err := update(Options{})
	//		if err != nil {
	//			log.Println("could not update state during rescale:", err.Error())
	//			return
	//		}
	//
	//		err = s.Open(context.Background())
	//		if err != nil {
	//			log.Println("could not open state during rescale:", err.Error())
	//		}
	//	}
	//
	// Otherwise, you are required to set this function yourself.
	// If you don't, New will panic.
	Rescale func(update func(Options) (*State, error))

	// ErrorHandler is the error handler of the event handler.
	//
	// Defaults to:
	//
	//	func(err error) {
	//		log.Println("event handler:", err.Error())
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

func (o *Options) setDefaults() error {
	if o.Token == "" {
		return errors.New("state: Options.Token may not be empty")
	}

	o.Token = "Bot " + o.Token

	if o.Status == "" {
		o.Status = discord.OnlineStatus
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
		o.Rescale = func(update func(Options) (*State, error)) {
			s, err := update(Options{})
			if err != nil {
				log.Println("could not update state during rescale:", err.Error())
				return
			}

			if err = s.Open(context.Background()); err != nil {
				log.Println("could not open state during rescale:", err.Error())
				return
			}
		}
	}

	if o.HTTPClient == nil {
		o.HTTPClient = httputil.NewClient()
	}

	if o.ErrorHandler == nil {
		o.ErrorHandler = func(err error) {
			log.Println("event handler:", err.Error())
		}
	}

	if o.PanicHandler == nil {
		o.PanicHandler = func(rec interface{}) {
			log.Printf("event handler: panic: %s\n", rec)
		}
	}

	return nil
}

// New creates a new *State using as many gateways as recommended by Discord.
func New(o Options) (*State, error) {
	if err := o.setDefaults(); err != nil {
		return nil, err
	}

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

	apiClient := api.NewCustomClient(o.Token, o.HTTPClient)
	ses := session.NewCustomSession(o.Gateways[0], apiClient, handler.New())

	s := &State{
		State:       state.NewFromSession(ses, o.Cabinet),
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
// Requesting user_ids will continue to be limited to returning 100 members.
func (s *State) RequestGuildMembers(d gateway.RequestGuildMembersData) error {
	return s.FromGuildID(d.GuildIDs[0]).RequestGuildMembers(d)
}

// onShardingRequired is the function stored as Gateway.OnShardingRequired
// in every of the Manager's gateways.
func (s *State) onShardingRequired() {
	if atomic.CompareAndSwapUint32(s.rescaleExec, 0, 1) {
		// make sure nobody can run apply
		s.mutex.Lock()
		defer s.mutex.Unlock()

		_ = s.Close()

		*s.rescaleExec = 0

		if s.rescale == nil {
			return
		}

		update := func(o Options) (*State, error) {
			if o.Token == "" {
				o.Token = s.options.Token
			}

			if o.Status == "" {
				o.Status = s.options.Status
			}

			if o.Activity == nil {
				o.Activity = s.options.Activity
			} else if o.Activity.Name == "" {
				o.Activity = nil
			}

			if o.Cabinet == nil {
				err := s.options.Cabinet.Reset()
				if err != nil {
					return nil, err
				}
				o.Cabinet = s.options.Cabinet
			}

			if o.Rescale == nil {
				o.Rescale = s.options.Rescale
			}

			newState, err := New(o)
			if err != nil {
				return nil, err
			}

			s.State = newState.State
			s.gateways = newState.gateways
			s.numShards = newState.numShards
			s.rescale = newState.rescale

			return s, nil
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
