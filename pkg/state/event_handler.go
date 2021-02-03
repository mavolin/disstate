package state

import (
	"errors"
	"reflect"
	"sync"

	"github.com/diamondburned/arikawa/v2/gateway"
)

var (
	// ErrInvalidHandler gets returned if a handler given to
	// EventManager.AddHandler or EventManager.MustAddHandler is not a valid
	// handler func, i.e. not following the form of func(*State, e) where e is
	// either a pointer to an event, *Base or interface{}.
	ErrInvalidHandler = errors.New("the passed interface{} does not resemble a valid handler")
	// ErrInvalidMiddleware gets returned if a middleware given to
	// EventManger.AddHandler or EventManager.MustAddHandler has not the same
	// type as its handler.
	//
	// Additionally, it is returned by AddGlobalMiddleware and
	// MustAddGlobalMiddleware if the middleware func is invalid.
	ErrInvalidMiddleware = errors.New("the passed middleware does not match the type of the handler")

	// Filtered should be returned if a filter blocks an event.
	Filtered = errors.New("filtered") //nolint:golint,stylecheck
)

var (
	interfaceType = reflect.TypeOf(func(interface{}) {}).In(0)
	baseType      = reflect.TypeOf(new(Base))
)

type (
	EventHandler struct {
		s  *State
		sv reflect.Value

		handlers      map[reflect.Type][]*genericHandler
		handlersMutex sync.RWMutex

		globalMiddlewares      map[reflect.Type][]globalMiddleware
		globalMiddlewaresMutex sync.RWMutex

		wg sync.WaitGroup

		ErrorHandler func(err error)
		PanicHandler func(err interface{})

		// currentSerial the next available serial number.
		// This is used to preserve the order of global middlewares.
		currentSerial uint64

		closer chan<- struct{}
	}

	globalMiddleware struct {
		middleware reflect.Value
		serial     uint64
	}

	// genericHandler wraps an event handler alongside it's middlewares.
	genericHandler struct {
		handler reflect.Value

		channel bool

		once *sync.Once
		rm   func()

		// middlewares are the middlewares for the handler.
		middlewares []middleware
	}

	middleware struct {
		middleware reflect.Value
		typ        reflect.Type
	}
)

// NewEventHandler creates a new EventHandler.
func NewEventHandler(s *State) *EventHandler {
	// make sure state update is blocking
	s.State.Session.Handler.Synchronous = true

	return &EventHandler{
		s:                 s,
		sv:                reflect.ValueOf(s),
		handlers:          make(map[reflect.Type][]*genericHandler),
		globalMiddlewares: make(map[reflect.Type][]globalMiddleware),
		ErrorHandler:      func(error) {},
		PanicHandler:      func(interface{}) {},
	}
}

// Open starts listening for events until the returned closer function is
// called.
func (h *EventHandler) Open(events <-chan interface{}) {
	closer := make(chan struct{})
	h.closer = closer

	go func() {
		for {
			select {
			case <-closer:
				return
			case gatewayEvent := <-events:
				e := h.genEvent(gatewayEvent)
				if e == nil {
					break
				}

				// prevent premature closer between here and when the first handler is called
				h.wg.Add(1)

				h.s.Session.Call(gatewayEvent) // trigger state update

				go func() {
					h.Call(e)
					h.wg.Done()
				}()
			}
		}
	}()
}

// Close stops the event listener and blocks until all handlers have finished
// executing.
func (h *EventHandler) Close() {
	if h.closer != nil {
		close(h.closer)
		h.closer = nil

		h.wg.Wait()
	}
}

var (
	stateType = reflect.TypeOf(new(State))
	errorType = reflect.TypeOf((error)(nil))
)

// DeriveIntents derives the intents based on the event handlers and global
// middlewares that were added.
// Interface and Base handlers will not be taken into account.
//
// Note that this does not reflect the intents needed to enable caching for
// API calls made anywhere in code.
func (h *EventHandler) DeriveIntents() (i gateway.Intents) {
	h.globalMiddlewaresMutex.RLock()

	for t := range h.globalMiddlewares {
		i |= eventIntents[t]
	}

	h.globalMiddlewaresMutex.RUnlock()
	h.handlersMutex.RLock()

	for t := range h.handlers {
		i |= eventIntents[t]
	}

	h.handlersMutex.RUnlock()
	return
}

// AddHandler adds a handler with the passed middlewares to the event handlers.
// A handler can either be a function, or a channel of type chan *eventType.
// Note, however, that channel sends are non-blocking, and you must either
// buffer your channel sufficiently, or ensure you are listening.
//
// The signature of a handler func is func(*State, e) where e is either a
// pointer to an event, *Base or interface{}.
// Optionally, a handler may return an error.
//
// Middlewares must be of the same type as the handlers or must be an
// interface{} or Base handlers.
func (h *EventHandler) AddHandler(handler interface{}, middlewares ...interface{}) (rm func(), err error) {
	return h.addHandler(handler, false, middlewares...)
}

func (h *EventHandler) addHandler(
	handler interface{}, execOnce bool, middlewares ...interface{},
) (rm func(), err error) {
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()

	var eventType reflect.Type

	if handlerType.Kind() == reflect.Chan {
		eventType = handlerType.Elem()
	} else if handlerType.Kind() == reflect.Func {
		// we expect two input params, first must be state
		if handlerType.NumIn() != 2 || handlerType.In(0) != stateType {
			return nil, ErrInvalidHandler
			// we expect either no return or an error return
		} else if (handlerType.NumOut() == 1 && handlerType.Out(1) != errorType) ||
			handlerType.NumOut() != 0 {
			return nil, ErrInvalidHandler
		}

		eventType = handlerType.In(1)
	} else {
		return nil, ErrInvalidHandler
	}

	gh := &genericHandler{
		handler: handlerVal,
		channel: handlerType.Kind() == reflect.Chan,
	}

	gh.middlewares, err = h.extractMiddlewares(middlewares, eventType)
	if err != nil {
		return nil, err
	}

	var once sync.Once

	rm = func() {
		once.Do(func() {
			h.handlersMutex.Lock()

			handler := h.handlers[handlerType]

			for i, ha := range handler {
				if ha == gh {
					h.handlers[handlerType] = append(handler[:i], handler[i+1:]...)
					break
				}
			}

			h.handlersMutex.Unlock()
		})
	}

	if execOnce {
		gh.once = new(sync.Once)
		gh.rm = rm
	}

	h.handlersMutex.Lock()
	h.handlers[eventType] = append(h.handlers[eventType], gh)
	h.handlersMutex.Unlock()

	return rm, nil
}

func (h *EventHandler) extractMiddlewares(raw []interface{}, eventType reflect.Type) ([]middleware, error) {
	mw := make([]middleware, len(raw))

	for i, m := range raw {
		mv := reflect.ValueOf(m)
		mt := mv.Type()

		if mt.Kind() != reflect.Func {
			return nil, ErrInvalidMiddleware
		}

		// we expect two input params, first must be state
		if mt.NumIn() != 2 || mt.In(0) != stateType {
			return nil, ErrInvalidMiddleware
			// we expect either no return or an error return
		} else if (mt.NumOut() == 1 && mt.Out(1) != errorType) || mt.NumOut() != 0 {
			return nil, ErrInvalidMiddleware
		}

		switch met := mt.In(1); met {
		case interfaceType, baseType, eventType:
			mw[i] = middleware{
				middleware: mv,
				typ:        met,
			}
		default:
			return nil, ErrInvalidMiddleware
		}
	}

	return mw, nil
}

// MustAddHandler is the same as AddHandler, but panics if AddHandler returns
// an error.
func (h *EventHandler) MustAddHandler(handler interface{}, middlewares ...interface{}) func() {
	r, err := h.AddHandler(handler, middlewares...)
	if err != nil {
		panic(err)
	}

	return r
}

// AddHandlerOnce adds a handler that is only executed once.
// If middlewares prevent execution, the handler will be executed on the next
// event.
func (h *EventHandler) AddHandlerOnce(handler interface{}, middlewares ...interface{}) error {
	_, err := h.addHandler(handler, true, middlewares...)
	return err
}

// MustAddHandlerOnce is the same as AddHandlerOnce, but panics if
// AddHandlerOnce returns an error.
func (h *EventHandler) MustAddHandlerOnce(handler interface{}, middlewares ...interface{}) {
	err := h.AddHandlerOnce(handler, middlewares...)
	if err != nil {
		panic(err)
	}
}

// AutoAddHandlers adds all handlers methods of the passed struct to the
// EventHandler.
// scan must be a pointer to a struct.
func (h *EventHandler) AutoAddHandlers(scan interface{}, middlewares ...interface{}) {
	v := reflect.ValueOf(scan)

	if v.Kind() != reflect.Ptr {
		return
	}

	for i := 0; i < v.NumMethod(); i++ {
		m := v.Method(i)

		if m.CanInterface() {
			// just try, AddHandler will abort if m is not a valid
			// handler func
			_, _ = h.AddHandler(m.Interface(), middlewares...)
		}
	}
}

// AddGlobalMiddleware adds the passed middleware as a global middleware.
//
// The scheme of a middleware func is func(*State, e) where e is either a
// pointer to an event, *Base or interface{}.
// Optionally, a middleware may return an error.
func (h *EventHandler) AddGlobalMiddleware(f interface{}) error {
	fv := reflect.ValueOf(f)
	ft := fv.Type()

	// we expect two input params, first must be state
	if ft.NumIn() != 2 || ft.In(0) != stateType {
		return ErrInvalidMiddleware
		// we expect either no return or an error return
	} else if !((ft.NumOut() == 1 && ft.Out(0) == errorType) || ft.NumOut() == 0) {
		return ErrInvalidMiddleware
	}

	et := ft.In(1)

	h.globalMiddlewaresMutex.Lock()
	h.globalMiddlewares[et] = append(h.globalMiddlewares[et], globalMiddleware{
		middleware: fv,
		serial:     h.currentSerial,
	})

	h.currentSerial++

	h.globalMiddlewaresMutex.Unlock()

	return nil
}

// MustAddGlobalMiddleware is the same as AddGlobalMiddleware but panics if
// AddGlobalMiddleware returns an error.
func (h *EventHandler) MustAddGlobalMiddleware(f interface{}) {
	err := h.AddGlobalMiddleware(f)
	if err != nil {
		panic(err)
	}
}

// Call can be used to manually dispatch an event.
// For this to succeed, e must be a pointer to an event, and it's Base field
// must be set.
func (h *EventHandler) Call(e interface{}) {
	ev := reflect.ValueOf(e)
	et := reflect.TypeOf(e)

	abort := h.callGlobalMiddlewares(ev, et)

	ev = ev.Elem() // from now functions only take elem

	direct := false

	switch e := e.(type) {
	case *ReadyEvent:
		h.handleReady(e)
	case *GuildCreateEvent:
		specificEvent := h.handleGuildCreate(e)
		if !abort {
			sev := reflect.ValueOf(specificEvent)
			set := reflect.TypeOf(specificEvent)
			h.call(sev, set, false)
		}

		direct = true
	case *GuildDeleteEvent:
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
// ev must not be a pointer, however, et is expected to be the pointerized type
// of ev.
//
// direct specifies, whether or not interface and Base handlers should be
// called for the event as well.
func (h *EventHandler) call(ev reflect.Value, et reflect.Type, direct bool) {
	h.handlersMutex.RLock()
	defer h.handlersMutex.RUnlock()

	if !direct {
		h.callHandlers(ev, et, h.handlers[interfaceType])
		h.callHandlers(ev, et, h.handlers[baseType])
	}

	h.callHandlers(ev, et, h.handlers[et])
}

// callHandlers calls the passed slice of handlers using the passed event ev.
// ev must not be a pointer, however, et is expected to be the pointerized type
// of ev.
func (h *EventHandler) callHandlers(ev reflect.Value, et reflect.Type, handlers []*genericHandler) {
	h.wg.Add(len(handlers))

	for _, gh := range handlers {
		go func(gh *genericHandler) {
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

			h.wg.Done()
		}(gh)
	}
}

func (h *EventHandler) callHandler(gh *genericHandler, ev reflect.Value) {
	if gh.channel {
		gh.handler.TrySend(ev)
	} else {
		result := gh.handler.Call([]reflect.Value{h.sv, ev})
		h.handleResult(result)
	}
}

// callGlobalMiddlewares calls the global middlewares using the passed event
// ev.
// ev must be a pointer to the event, and et must be ev's type.
func (h *EventHandler) callGlobalMiddlewares(ev reflect.Value, et reflect.Type) bool {
	h.globalMiddlewaresMutex.RLock()

	interfaceMiddlewares := h.globalMiddlewares[interfaceType]
	baseMiddlewares := h.globalMiddlewares[baseType]
	typedMiddlewares := h.globalMiddlewares[et]

	h.globalMiddlewaresMutex.RUnlock()

	var im, bm, tm int

	for {
		var (
			next  globalMiddleware
			typ   reflect.Type
			index *int = nil
		)

		if im < len(interfaceMiddlewares) {
			next = interfaceMiddlewares[im]
			typ = et
			index = &im
		}

		if bm < len(baseMiddlewares) && (index == nil || baseMiddlewares[bm].serial < next.serial) {
			next = baseMiddlewares[bm]
			typ = baseType
			index = &bm
		}

		if tm < len(typedMiddlewares) && (index == nil || typedMiddlewares[tm].serial < next.serial) {
			next = typedMiddlewares[tm]
			typ = et
			index = &tm
		}

		if index == nil {
			break // every middleware consumed
		}

		var in2 reflect.Value

		switch typ {
		case et:
			in2 = ev
		case baseType:
			in2 = ev.Elem().FieldByName("Base")
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

			result = next.middleware.Call([]reflect.Value{h.sv, in2})
		}()

		if didPanic {
			return true
		}

		if h.handleResult(result) {
			return true
		}

		*index++
	}

	return false
}

// callMiddlewares calls the passed slice of middlewares in the passed order.
// ev must not be a pointer, however, et is expected to be the pointerized type
// of ev.
func (h *EventHandler) callMiddlewares(ev reflect.Value, et reflect.Type, middlewares []middleware) bool {
	for _, m := range middlewares {
		var (
			result []reflect.Value
			base   reflect.Value
		)

		switch m.typ {
		case interfaceType:
			result = m.middleware.Call([]reflect.Value{h.sv, ev})
		case baseType:
			if !base.IsValid() {
				base = ev.Elem().FieldByName("Base")
			}

			result = m.middleware.Call([]reflect.Value{h.sv, base})
		case et:
			result = m.middleware.Call([]reflect.Value{h.sv, ev})
		default: // skip invalid
			continue
		}

		if h.handleResult(result) {
			return true
		}
	}

	return false
}

func (h *EventHandler) handleReady(e *ReadyEvent) {
	for _, g := range e.Guilds {
		// store this so we know when we need to dispatch the corresponding
		// GuildReadyEvent
		h.s.unreadyGuilds.Add(g.ID)
	}
}

func (h *EventHandler) handleGuildCreate(e *GuildCreateEvent) interface{} {
	switch {
	// this guild was unavailable, but has come back online
	case h.s.unavailableGuilds.Delete(e.ID):
		return &GuildAvailableEvent{GuildCreateEvent: e}

	// the guild was announced in Ready and has now become available
	case h.s.unreadyGuilds.Delete(e.ID):
		return &GuildReadyEvent{GuildCreateEvent: e}

	// we don't know this guild, hence we just joined it
	default:
		return &GuildJoinEvent{GuildCreateEvent: e}
	}
}

func (h *EventHandler) handleGuildDelete(e *GuildDeleteEvent) interface{} {
	// store this so we can later dispatch a GuildAvailableEvent, once the
	// guild becomes available again.
	if e.Unavailable {
		h.s.unavailableGuilds.Add(e.ID)

		return &GuildUnavailableEvent{GuildDeleteEvent: e}
	}

	// it might have been unavailable before we left
	h.s.unavailableGuilds.Delete(e.ID)

	return &GuildLeaveEvent{GuildDeleteEvent: e}
}
