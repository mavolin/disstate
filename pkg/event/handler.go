package event

import (
	"errors"
	"reflect"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

var (
	// ErrInvalidHandler gets returned if a handler given to
	// Handler.AddHandler or Handler.MustAddHandler is not a valid handler func.
	ErrInvalidHandler = errors.New("state: the passed interface{} does not resemble a valid handler")
	// ErrInvalidMiddleware gets returned if a middleware given to
	// Handler.AddHandler or Handler.MustAddHandler is invalid.
	ErrInvalidMiddleware = errors.New("state: the passed middleware does not match the type of the handler")

	// Filtered should be returned if a filter blocks an event.
	Filtered = errors.New("filtered") //nolint:revive
)

type (
	// Handler is the event handler.
	//
	// For a detailed list of differences of arikawa and disstate's event
	// handling refer to the package doc.
	Handler struct {
		state  *state.State
		rstate reflect.Value

		globalMiddlewares map[reflect.Type][]globalMiddleware
		handlers          map[reflect.Type][]*handlerMeta
		mutex             sync.RWMutex

		ErrorHandler func(err error)
		PanicHandler func(err interface{})

		// currentSerial is the next available serial number.
		// This is used to preserve the order of global middlewares.
		currentSerial uint

		handlerWG sync.WaitGroup
		closer    chan<- struct{}

		// unavailableGuilds is a set of discord.GuildIDs of guilds that became
		// unavailable after connecting to the gateway, i.e. they were sent in
		// a GuildUnavailableEvent.
		unavailableGuilds map[discord.GuildID]struct{}
		// unreadyGuilds is a set of discord.GuildIDs of the guilds received
		// during the Ready event.
		// After receiving guild create events for those guilds, they will be
		// removed.
		unreadyGuilds map[discord.GuildID]struct{}
		guildMutex    sync.Mutex
	}

	globalMiddleware struct {
		middleware reflect.Value
		serial     uint
	}

	handlerMeta struct {
		handler reflect.Value

		// once is a *sync.Once used if the handler shall only be executed
		// once.
		once *sync.Once
		// rm is the function called if the handler shall only be invoked once
		// and the handler is therefore removed.
		rm func()

		// middlewares contains the per-handler middlewares.
		middlewares []reflect.Value
	}
)

// NewHandler creates a new Handler using the passed reflect.Value of the
// *state.State.
// It panics if the reflect.Value is not of the correct type.
func NewHandler(stateVal reflect.Value) *Handler {
	return &Handler{
		state:             stateVal.Elem().FieldByName("State").Interface().(*state.State),
		rstate:            stateVal,
		globalMiddlewares: make(map[reflect.Type][]globalMiddleware),
		handlers:          make(map[reflect.Type][]*handlerMeta),
		ErrorHandler:      func(error) {},
		PanicHandler:      func(interface{}) {},
		unavailableGuilds: make(map[discord.GuildID]struct{}),
		unreadyGuilds:     make(map[discord.GuildID]struct{}),
	}
}

// Open listens to events on the passed channel until Close is called.
func (h *Handler) Open(events <-chan interface{}) {
	closer := make(chan struct{})
	h.closer = closer

	go func() {
		for {
			select {
			case <-closer:
				return
			case gatewayEvent := <-events:
				e := h.generateEvent(gatewayEvent)
				if e == nil {
					break
				}

				// prevent premature closer between here and when the first handler is called
				h.handlerWG.Add(1)

				h.state.Session.Call(gatewayEvent) // trigger state update

				go func() {
					h.Call(e)
					h.handlerWG.Done()
				}()
			}
		}
	}()
}

// Close signals the event listener to stop and blocks until all handlers
// have finished their execution.
func (h *Handler) Close() {
	if h.closer != nil {
		close(h.closer)
		h.closer = nil

		h.handlerWG.Wait()
	}
}

// DeriveIntents derives the intents based on the event handlers and global
// middlewares that were added.
// Interface and Base handlers will not be taken into account.
//
// Note that this does not reflect the intents needed to enable caching for
// API calls made anywhere in code.
func (h *Handler) DeriveIntents() (i gateway.Intents) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for t := range h.globalMiddlewares {
		i |= eventIntents[t]
	}

	for t := range h.handlers {
		i |= eventIntents[t]
	}

	return
}

// TryAddHandler is the same as AddHandler, but returns an error if the
// signature of the handler or one of the middlewares is invalid.
func (h *Handler) TryAddHandler(handler interface{}, middlewares ...interface{}) (rm func(), err error) {
	return h.addHandler(handler, false, middlewares...)
}

// AddHandler adds a handler with the passed middlewares to the event handler.
//
// The handler can either be a channel of pointer to an event type, or a
// function.
// Note that the event handler will not wait until the channel is ready to
// receive.
// Instead you must ensure, that your channel is sufficiently buffered or
// you are ready to receive.
// If you require otherwise, consider add a handler function that send to your
// channel blockingly.
//
// If using a function as handler, the function must match func(*State, e),
// where e is either a pointer to an event, *Base, or interface{}.
// Optionally, a handler function may return an error that will be handled by
// Handler.ErrorHandler, if non-nil.
//
// The same requirements as for functions apply to middlewares as well.
func (h *Handler) AddHandler(handler interface{}, middlewares ...interface{}) func() {
	rm, err := h.TryAddHandler(handler, middlewares...)
	if err != nil {
		panic(err)
	}

	return rm
}

// TryAddHandlerOnce is the same as AddHandlerOnce, but returns an error if
// the signature of the handler of of one of the middlewares is invalid.
func (h *Handler) TryAddHandlerOnce(handler interface{}, middlewares ...interface{}) error {
	_, err := h.addHandler(handler, true, middlewares...)
	return err
}

// AddHandlerOnce is the same as AddHandler, but only handles a single event
// before removing itself.
func (h *Handler) AddHandlerOnce(handler interface{}, middlewares ...interface{}) {
	err := h.TryAddHandlerOnce(handler, middlewares...)
	if err != nil {
		panic(err)
	}
}

// AutoAddHandlers adds all handlers methods of the passed struct to the
// Handler.
// scan must be a pointer to a struct.
func (h *Handler) AutoAddHandlers(scan interface{}, middlewares ...interface{}) {
	v := reflect.ValueOf(scan)

	if v.Kind() != reflect.Ptr {
		return
	}

	for i := 0; i < v.NumMethod(); i++ {
		m := v.Method(i)

		if m.CanInterface() {
			// just try, TryAddHandler will abort if m is not a valid handler
			// func
			_, _ = h.TryAddHandler(m.Interface(), middlewares...)
		}
	}
}

func (h *Handler) addHandler(handler interface{}, execOnce bool, middlewares ...interface{}) (rm func(), err error) {
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()

	var eventType reflect.Type

	switch handlerType.Kind() {
	case reflect.Chan:
		eventType = handlerType.Elem()
	case reflect.Func:
		// we expect two input params, first must be state
		if handlerType.NumIn() != 2 || handlerType.In(0) != h.rstate.Type() {
			return nil, ErrInvalidHandler
			// we expect either no return or an error return
		} else if handlerType.NumOut() != 0 && (handlerType.NumOut() != 1 || handlerType.Out(0) != errorType) {
			return nil, ErrInvalidHandler
		}

		eventType = handlerType.In(1)
	default:
		return nil, ErrInvalidHandler
	}

	gh := &handlerMeta{
		handler: handlerVal,
	}

	gh.middlewares, err = h.extractMiddlewares(middlewares, eventType)
	if err != nil {
		return nil, err
	}

	var once sync.Once

	rm = func() {
		once.Do(func() {
			h.mutex.Lock()
			defer h.mutex.Unlock()

			handler := h.handlers[handlerType]

			for i, ha := range handler {
				if ha == gh {
					h.handlers[handlerType] = append(handler[:i], handler[i+1:]...)
					break
				}
			}
		})
	}

	if execOnce {
		gh.once = new(sync.Once)
		gh.rm = rm
	}

	h.mutex.Lock()
	h.handlers[eventType] = append(h.handlers[eventType], gh)
	h.mutex.Unlock()

	return rm, nil
}

func (h *Handler) extractMiddlewares(raw []interface{}, eventType reflect.Type) ([]reflect.Value, error) {
	mw := make([]reflect.Value, len(raw))

	for i, m := range raw {
		mv := reflect.ValueOf(m)
		mt := mv.Type()

		if mt.Kind() != reflect.Func {
			return nil, ErrInvalidMiddleware
		}

		// we expect two input params, first must be state
		if mt.NumIn() != 2 || mt.In(0) != h.rstate.Type() {
			return nil, ErrInvalidMiddleware
			// we expect either no return or an error return
		} else if mt.NumOut() != 0 && (mt.NumOut() != 1 || mt.Out(0) != errorType) {
			return nil, ErrInvalidMiddleware
		}

		switch met := mt.In(1); met {
		case interfaceType, baseType, eventType:
			mw[i] = mv
		default:
			return nil, ErrInvalidMiddleware
		}
	}

	return mw, nil
}

// AddMiddleware adds the passed middleware as a global middleware.
//
// The signature of a middleware func is func(*State, e) where e is either a
// pointer to an event, *Base or interface{}.
// Optionally, a middleware may return an error.
func (h *Handler) AddMiddleware(f interface{}) {
	if err := h.TryAddMiddleware(f); err != nil {
		panic(err)
	}
}

// TryAddMiddleware is the same as AddMiddleware, but returns an error if
// the signature of the middleware is invalid.
func (h *Handler) TryAddMiddleware(f interface{}) error {
	fv := reflect.ValueOf(f)
	ft := fv.Type()

	// we expect two input params, first must be state
	if ft.NumIn() != 2 || ft.In(0) != h.rstate.Type() {
		return ErrInvalidMiddleware
		// we expect either no return or an error return
	} else if ft.NumOut() != 0 && (ft.NumOut() != 1 || ft.Out(0) != errorType) {
		return ErrInvalidMiddleware
	}

	et := ft.In(1)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.globalMiddlewares[et] = append(h.globalMiddlewares[et], globalMiddleware{
		middleware: fv,
		serial:     h.currentSerial,
	})

	h.currentSerial++

	return nil
}

// Call can be used to manually dispatch an event.
// For this to succeed, e must be a pointer to an event, and it's Base field
// must be set.
func (h *Handler) Call(e interface{}) {
	ev := reflect.ValueOf(e)
	et := reflect.TypeOf(e)

	abort := h.callGlobalMiddlewares(ev, et)
	var direct bool

	switch e := e.(type) {
	case *Ready:
		h.handleReady(e)
	case *GuildCreate:
		specificEvent := h.handleGuildCreate(e)
		if !abort {
			sev := reflect.ValueOf(specificEvent)
			set := reflect.TypeOf(specificEvent)
			h.call(sev, set, false)
		}

		direct = true
	case *GuildDelete:
		specificEvent := h.handleGuildDelete(e)
		if !abort {
			sev := reflect.ValueOf(specificEvent)
			set := reflect.TypeOf(specificEvent)
			h.call(sev, set, false)
		}

		direct = true
	}

	if !abort {
		h.call(ev, et, direct)
	}
}

// call calls the handlers for the passed typed using the event wrapped in ev.
//
// direct specifies, whether interface and Base handlers should be called for
// the event as well.
func (h *Handler) call(ev reflect.Value, et reflect.Type, direct bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if !direct {
		h.callHandlers(ev, et, h.handlers[interfaceType])
		h.callHandlers(ev, et, h.handlers[baseType])
	}

	h.callHandlers(ev, et, h.handlers[et])
}

// callHandlers calls the passed slice of handlers using the passed event ev.
func (h *Handler) callHandlers(ev reflect.Value, et reflect.Type, handlers []*handlerMeta) {
	h.handlerWG.Add(len(handlers))

	for _, gh := range handlers {
		go func(gh *handlerMeta) {
			defer h.handlerWG.Done()

			defer func() {
				if rec := recover(); rec != nil {
					h.PanicHandler(rec)
				}
			}()

			cp := copyEvent(ev, et)

			if h.callMiddlewares(cp, et, gh.middlewares) {
				return
			}

			if gh.once != nil {
				gh.once.Do(func() {
					h.callHandler(gh, cp)
					gh.rm()
				})
			} else {
				h.callHandler(gh, cp)
			}
		}(gh)
	}
}

func (h *Handler) callHandler(gh *handlerMeta, ev reflect.Value) {
	if gh.handler.Type().Kind() == reflect.Chan {
		gh.handler.TrySend(ev)
	} else {
		result := gh.handler.Call([]reflect.Value{h.rstate, ev})
		h.handleResult(result)
	}
}

// callGlobalMiddlewares calls the global middlewares using the passed event
// ev.
func (h *Handler) callGlobalMiddlewares(ev reflect.Value, et reflect.Type) bool {
	h.mutex.RLock()

	interfaceMiddlewares := h.globalMiddlewares[interfaceType]
	baseMiddlewares := h.globalMiddlewares[baseType]
	typedMiddlewares := h.globalMiddlewares[et]

	h.mutex.RUnlock()

	var im, bm, tm int

	for {
		var (
			next  globalMiddleware
			typ   reflect.Type
			index *int
		)

		// if there are interface middlewares left, use it to compare against
		if im < len(interfaceMiddlewares) {
			next = interfaceMiddlewares[im]
			typ = et
			index = &im
		}

		// if there are base middlewares and there is no interface middleware,
		// or this middleware was added before the interface middleware,
		// select it as next middleware
		if bm < len(baseMiddlewares) && (index == nil || baseMiddlewares[bm].serial < next.serial) {
			next = baseMiddlewares[bm]
			typ = baseType
			index = &bm
		}

		// if there are typed middlewares and there is no interface- and no
		// base middleware or this middleware was added before both, select it
		// as next middleware
		if tm < len(typedMiddlewares) && (index == nil || typedMiddlewares[tm].serial < next.serial) {
			next = typedMiddlewares[tm]
			typ = et
			index = &tm
		}

		// if we found no next middleware, i.e. we consumed all
		if index == nil {
			break
		}

		// increase the index for the middleware slice we took our next
		// middleware from
		*index++
		// and reset index for our next iteration
		index = nil

		var in1 reflect.Value

		switch typ {
		case et:
			in1 = ev
		case baseType:
			in1 = ev.Elem().FieldByName("Base")
		default:
			continue // invalid, skip
		}

		var (
			result   []reflect.Value
			didPanic bool
		)

		func() {
			defer func() {
				if rec := recover(); rec != nil {
					h.PanicHandler(rec)
					didPanic = true
				}
			}()

			result = next.middleware.Call([]reflect.Value{h.rstate, in1})
		}()

		if didPanic {
			return true
		}

		if h.handleResult(result) {
			return true
		}
	}

	return false
}

// callMiddlewares calls the passed slice of middlewares in the passed order.
func (h *Handler) callMiddlewares(ev reflect.Value, et reflect.Type, middlewares []reflect.Value) bool {
	var base reflect.Value

	for _, m := range middlewares {
		var result []reflect.Value

		switch m.Type().In(1) {
		case interfaceType, et:
			result = m.Call([]reflect.Value{h.rstate, ev})
		case baseType:
			if !base.IsValid() {
				base = ev.Elem().FieldByName("Base")
			}

			result = m.Call([]reflect.Value{h.rstate, base})
		default: // skip invalid
			continue
		}

		if h.handleResult(result) {
			return true
		}
	}

	return false
}

func (h *Handler) handleReady(e *Ready) {
	h.guildMutex.Lock()
	defer h.guildMutex.Unlock()

	for _, g := range e.Guilds {
		h.unreadyGuilds[g.ID] = struct{}{}
	}
}

func (h *Handler) handleGuildCreate(e *GuildCreate) interface{} {
	h.guildMutex.Lock()
	defer h.guildMutex.Unlock()

	// The guild was previously announced to us in the ready event, and has now
	// become available.
	if _, ok := h.unreadyGuilds[e.ID]; ok {
		delete(h.unreadyGuilds, e.ID)
		return &GuildReady{GuildCreate: e}

		// The guild was previously announced as unavailable through a guild
		// delete event, and has now become available again.
	} else if _, ok = h.unavailableGuilds[e.ID]; ok {
		delete(h.unavailableGuilds, e.ID)
		return &GuildAvailable{GuildCreate: e}
	}

	// We don't know this guild, hence it's new.
	return &GuildJoin{GuildCreate: e}
}

func (h *Handler) handleGuildDelete(e *GuildDelete) interface{} {
	h.guildMutex.Lock()
	defer h.guildMutex.Unlock()

	// store this so we can later dispatch a GuildAvailableEvent, once the
	// guild becomes available again.
	if e.Unavailable {
		h.unavailableGuilds[e.ID] = struct{}{}

		return &GuildUnavailable{GuildDelete: e}
	}

	// Possible scenario requiring this would be leaving the guild while
	// unavailable.
	delete(h.unavailableGuilds, e.ID)

	return &GuildLeave{GuildDelete: e}
}

// handleResult handles the passed result of a handler func.
func (h *Handler) handleResult(res []reflect.Value) bool {
	if len(res) == 0 {
		return false
	}

	err := res[0].Interface().(error)
	if errors.Is(err, Filtered) {
		return true
	} else if err != nil {
		h.ErrorHandler(err)
		return true
	}

	return false
}

// copyEvent copies the event stored in the passed reflect.Value with the
// passed reflect.Type.
// v must not be a pointer however, t is expected to be the pointerized type
// of v.
func copyEvent(v reflect.Value, t reflect.Type) reflect.Value {
	v = v.Elem()

	cp := reflect.New(t.Elem())
	cp = cp.Elem()

	for i := 0; i < v.NumField(); i++ {
		cp.Field(i).Set(v.Field(i))
	}

	b := v.FieldByName("Base").Interface().(*Base)
	bcp := b.copy()

	bcpValue := reflect.ValueOf(bcp)

	cp.FieldByName("Base").Set(bcpValue)

	return cp.Addr()
}
