package state

import (
	"errors"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

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
	// To remove the current activity when calling update during rescale, set
	// this to an empty activity.
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
	//
	// Default: 0..TotalShards
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
	// To update the state's shard manager, you must call update.
	// All zero-value options in the Options you give to update, will be set to
	// the options you used when initially creating the state.
	// This does not apply to TotalShards, ShardIDs, and Gateways, which will
	// assume the defaults described in their respective documentation.
	// Furthermore, setting ErrorHandler or PanicHandler will have no effect.
	//
	// After calling update, you should reopen the state, by calling Open.
	// Alternatively, you can call open individually for State.Gateways().
	// Note, however, that you should call Sate.Handler.Open(State.Events) once
	// before calling Gateway.Open, should you choose to open individually.
	//
	// During update, the state's State field will be replaced, as well as the
	// gateways and the rescale function. The event handler will remain
	// untouched which is why you don't need to re-add your handlers.
	//
	// Default
	//
	// If you set neither TotalShards nor Gateways, this will default to the
	// below unless you define a custom Rescale function.
	//
	// 	func(update func(Options) *State) {
	//		s, err := update(Options{})
	//		if err != nil {
	//			log.Println("could not update state during rescale:", err.Error())
	//			return
	//		}
	//
	//		err = s.Open(2*time.Second))
	//		if err != nil {
	//			log.Println("could not open state during rescale:", err.Error())
	//		}
	//	}
	//
	// Otherwise, you are required to set this function yourself.
	Rescale func(update func(Options) (*State, error))

	// ErrorHandler is the error handler of the event handler.
	//
	// Default
	//
	//	func(err error) {
	//		log.Println("event handler:", err.Error())
	//	}
	ErrorHandler func(error)
	// PanicHandler is the panic handler of the event handler
	//
	// Default
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
		return errors.New("state: setting Options.ShardIDs requires Options.TotalShards to be set as well")
	}

	if o.TotalShards > 0 && o.Rescale == nil {
		return errors.New("state: setting Options.TotalShards requires Options.Rescale to be set as well")
	}

	if o.TotalShards <= 0 && o.Rescale == nil && len(o.Gateways) == 0 {
		o.Rescale = func(update func(Options) (*State, error)) {
			s, err := update(Options{})
			if err != nil {
				log.Println("could not update state during rescale:", err.Error())
				return
			}

			if err = s.Open(2 * time.Second); err != nil {
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
