package state

import (
	"errors"
	"reflect"
	"sync"
)

var (
	// ErrInvalidHandler gets returned if a handler given to
	// EventManager.AddHandler or EventManager.MustAddHandler is not a valid
	// handlers func, i.e. not following the form of func(*State, e) where e is
	// either a pointer to an event, *Base or interface{}.
	ErrInvalidHandler = errors.New("the passed interface{} does not resemble a valid handlers")
	// ErrInvalidMiddleware gets returned if a middleware given to
	// EventManger.AddHandler or EventManager.MustAddHandler has not the same
	// type as its handler.
	//
	// Additionally, it is returned by AddGlobalMiddleware and
	// MustAddGlobalMiddleware if the middleware func is invalid.
	ErrInvalidMiddleware = errors.New("the passed middleware does not match the type of the handlers")

	// Filtered should be returned if a filter blocks an event.
	Filtered = errors.New("filtered")
)

var (
	interfaceType = reflect.TypeOf(func(*State, interface{}) error { return nil })
	baseType      = reflect.TypeOf(func(*State, *Base) error { return nil })
)

type EventHandler struct {
	s  *State
	sv reflect.Value

	handlers      map[reflect.Type][]*genericHandler
	handlersMutex sync.RWMutex

	globalMiddlewares      map[reflect.Type][]reflect.Value
	globalMiddlewaresMutex sync.RWMutex

	ErrorHandler func(err error)
	PanicHandler func(err interface{})

	closer chan<- struct{}
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(s *State) *EventHandler {
	return &EventHandler{
		s:                 s,
		sv:                reflect.ValueOf(s),
		handlers:          make(map[reflect.Type][]*genericHandler),
		globalMiddlewares: make(map[reflect.Type][]reflect.Value),
		ErrorHandler:      func(error) {},
		PanicHandler:      func(interface{}) {},
	}
}

// genericHandler wraps an event handler.
type genericHandler struct {
	handler reflect.Value

	// middlewares are the middlewares for the handler.
	interfaceMiddlewares []reflect.Value
	baseMiddlewares      []reflect.Value
	typedMiddlewares     []reflect.Value
}

// Open starts listening for events until the returned closer function is
// called.
func (h *EventHandler) Open(events <-chan interface{}) {
	closer := make(chan struct{})

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

				h.s.updateStore(gatewayEvent)
				go h.Call(e)
			}
		}
	}()
}

func (h *EventHandler) Close() {
	if h.closer != nil {
		h.closer <- struct{}{}
	}
}

var (
	stateType = reflect.TypeOf(new(State))
	errorType = reflect.TypeOf(func(error) {}).In(0)
)

// AddHandler adds a handlers with the passed globalMiddlewares to the event handlers.
//
// Middlewares must be of the same type as the handlers or must be an
// interface{} or Base handlers.
//
// The scheme of a handler func is func(*State, e) where e is either a pointer
// to an event, *Base or interface{}.
// Optionally, a handler may return an error.
func (h *EventHandler) AddHandler(f interface{}, middlewares ...interface{}) (func(), error) {
	fv := reflect.ValueOf(f)
	ft := fv.Type()

	// we expect two input params, first must be state
	if ft.NumIn() != 2 || ft.In(0) != stateType {
		return nil, ErrInvalidHandler
		// we expect either no return or an error return
	} else if (ft.NumOut() == 1 && ft.Out(1) != errorType) || ft.NumOut() != 0 {
		return nil, ErrInvalidHandler
	}

	gh := &genericHandler{
		handler:          fv,
		typedMiddlewares: make([]reflect.Value, 0, len(middlewares)),
	}

	for _, m := range middlewares {
		mv := reflect.ValueOf(m)
		mt := mv.Type()

		if mt == interfaceType {
			gh.interfaceMiddlewares = append(gh.interfaceMiddlewares, mv)
		} else if mt == baseType {
			gh.baseMiddlewares = append(gh.baseMiddlewares, mv)
		} else if mt == ft {
			gh.typedMiddlewares = append(gh.typedMiddlewares, mv)
		} else {
			return nil, ErrInvalidMiddleware
		}
	}

	et := ft.In(1)

	h.handlersMutex.Lock()
	h.handlers[et] = append(h.handlers[et], gh)
	h.handlersMutex.Unlock()

	var once sync.Once

	return func() {
		once.Do(func() {
			h.handlersMutex.Lock()

			handler := h.handlers[ft]

			for i, ha := range handler {
				if ha == gh {
					h.handlers[ft] = append(handler[:i], handler[i+1:]...)
					break
				}
			}

			h.handlersMutex.Unlock()
		})
	}, nil
}

// MustAddHandler is the same as AddHandler, but panics if AddHandler returns
// an error.
func (h *EventHandler) MustAddHandler(f interface{}, middlewares ...interface{}) func() {
	r, err := h.AddHandler(f, middlewares...)
	if err != nil {
		panic(err)
	}

	return r
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
			// we try to add, AddHandler will abort if m is not a valid handlers
			// func
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
	} else if (ft.NumOut() == 1 && ft.Out(1) != errorType) || ft.NumOut() != 0 {
		return ErrInvalidMiddleware
	}

	et := ft.In(1)

	h.globalMiddlewaresMutex.Lock()
	h.globalMiddlewares[et] = append(h.globalMiddlewares[et], fv)
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
		if specificEvent := h.handleGuildCreate(e); !abort {
			sev := reflect.ValueOf(specificEvent)
			set := reflect.TypeOf(specificEvent)
			h.call(sev, set, false)
		}

		direct = true
	case *GuildDeleteEvent:
		if specificEvent := h.handleGuildDelete(e); !abort {
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
func (h *EventHandler) callHandlers(ev reflect.Value, et reflect.Type, gh []*genericHandler) {
	for _, handler := range gh {
		go func(ha *genericHandler) {
			defer func() {
				if rec := recover(); rec != nil {
					h.PanicHandler(rec)
				}
			}()

			cp := copyEvent(ev, et)

			if h.callMiddlewares(cp, ha.interfaceMiddlewares, ha.baseMiddlewares, ha.typedMiddlewares) {
				return
			}

			result := ha.handler.Call([]reflect.Value{h.sv, cp})
			h.handleResult(result)
		}(handler)
	}
}

// callGlobalMiddlewares calls the global middlewares using the passed event
// ev.
// ev must be a pointer to the event, and et must be ev's type.
func (h *EventHandler) callGlobalMiddlewares(ev reflect.Value, et reflect.Type) bool {
	h.globalMiddlewaresMutex.RLock()
	defer h.globalMiddlewaresMutex.RUnlock()

	for _, m := range h.globalMiddlewares[interfaceType] {
		result := m.Call([]reflect.Value{h.sv, ev})
		if h.handleResult(result) {
			return true
		}
	}

	base := ev.Elem().FieldByName("Base")

	for _, m := range h.globalMiddlewares[baseType] {
		result := m.Call([]reflect.Value{h.sv, base})
		if h.handleResult(result) {
			return true
		}
	}

	for _, m := range h.globalMiddlewares[et] {
		result := m.Call([]reflect.Value{h.sv, ev})
		if h.handleResult(result) {
			return true
		}
	}

	return false
}

// callMiddlewares calls the passed slices of middlewares in the slices order.
// ev must not be a pointer, however, et is expected to be the pointerized type
// of ev.
func (h *EventHandler) callMiddlewares(
	ev reflect.Value, interfaceMiddlewares, baseMiddlewares, typedMiddlewares []reflect.Value,
) bool {
	for _, m := range interfaceMiddlewares {
		result := m.Call([]reflect.Value{h.sv, ev})
		if h.handleResult(result) {
			return true
		}
	}

	base := ev.Elem().FieldByName("Base")

	for _, m := range baseMiddlewares {
		result := m.Call([]reflect.Value{h.sv, base})
		if h.handleResult(result) {
			return true
		}
	}

	for _, m := range typedMiddlewares {
		result := m.Call([]reflect.Value{h.sv, ev})
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
	// this guild was unavailable, but has come back online
	if h.s.unavailableGuilds.Delete(e.ID) {
		return &GuildAvailableEvent{
			GuildCreateEvent: e,
		}

		// the guild was announced in Ready and has now become available
	} else if h.s.unreadyGuilds.Delete(e.ID) {
		return &GuildReadyEvent{
			GuildCreateEvent: e,
		}
	} else { // we don't know this guild, hence we just joined it
		return &GuildJoinEvent{
			GuildCreateEvent: e,
		}
	}
}

func (h *EventHandler) handleGuildDelete(e *GuildDeleteEvent) interface{} {
	// store this so we can later dispatch a GuildAvailableEvent, once the
	// guild becomes available again.
	if e.Unavailable {
		h.s.unavailableGuilds.Add(e.ID)

		return &GuildUnavailableEvent{
			GuildDeleteEvent: e,
		}
	} else {
		// it might have been unavailable before we left
		h.s.unavailableGuilds.Delete(e.ID)

		return &GuildLeaveEvent{
			GuildDeleteEvent: e,
		}
	}
}
