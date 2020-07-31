package state

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

var (
	// ErrNotAHandler gets returned if a handler given to one of the
	// EventManager.HandleX methods is not a valid handler, i.e. following
	// the form of func(State, e where e is either a pointer to an
	// event or interface{}, or if the value is not a function at all.
	ErrNotAHandler = errors.New("the passed interface{} does not resemble a valid handler")
	// ErrNotAMiddleware gets returned if a middleware given to
	// EventManger.AddMiddleware is not a valid middleware, i.e. following the
	// form of func(State, e where e is either a pointer to an event or
	// interface{}, or if the value is not a function at all.
	ErrNotAMiddleware = errors.New("the passed interface{} does not resemble a valid middleware")

	// Filtered should be returned if a filter blocks an event.
	Filtered = errors.New("filtered")
)

type EventHandler struct {
	s *State

	handler      map[eventType][]*genericHandler
	handlerMutex sync.RWMutex

	middlewares      map[eventType][]handler
	middlewaresMutex sync.RWMutex

	ErrorHandler func(err error)
	PanicHandler func(err interface{})
}

func NewEventHandler(s *State) *EventHandler {
	return &EventHandler{
		s:            s,
		handler:      make(map[eventType][]*genericHandler),
		middlewares:  make(map[eventType][]handler),
		ErrorHandler: func(error) {},
		PanicHandler: func(interface{}) {},
	}
}

// genericHandler wraps an event handler.
type genericHandler struct {
	// handler is the underlying handler
	handler handler
	// filtered specifies whether this event is filter, i.e. it will not be
	// fired if a filter blocks it.
	filtered bool
}

func (h *EventHandler) Open(events <-chan interface{}) chan<- struct{} {
	closer := make(chan struct{})

	go func() {
		for {
			select {
			case <-closer:
				return
			case e := <-events:
				de, b, t := h.genEvent(e)
				if e == nil {
					break
				}

				h.s.updateStore(e)
				h.call(de, b, t)
			}
		}
	}()

	return closer
}

func (h *EventHandler) AddHandler(f interface{}) func() {
	return h.addHandler(f, true)
}

func (h *EventHandler) AddUnfilteredHandler(f interface{}) func() {
	return h.addHandler(f, false)
}

func (h *EventHandler) addHandler(f interface{}, filtered bool) func() {
	ha, t := handlerFuncForHandler(f)
	if ha == nil {
		panic(ErrNotAHandler)
	}

	gh := &genericHandler{
		handler:  ha,
		filtered: filtered,
	}

	h.handlerMutex.Lock()
	h.handler[t] = append(h.handler[t], gh)
	h.handlerMutex.Unlock()

	return func() {
		h.handlerMutex.Lock()

		handler := h.handler[t]

		for i, ha := range handler {
			if ha == gh {
				h.handler[t] = append(handler[:i], handler[i+1:]...)
				break
			}
		}

		h.handlerMutex.Unlock()
	}
}

func (h *EventHandler) AddMiddleware(f interface{}) {
	ha, t := handlerFuncForHandler(f)
	if ha == nil {
		panic(ErrNotAMiddleware)
	}

	h.middlewaresMutex.Lock()
	h.middlewares[t] = append(h.middlewares[t], ha)
	h.middlewaresMutex.Unlock()
}

func (h *EventHandler) Call(e interface{}) {
	t := calcEventType(e)
	if t == 0 {
		return
	}

	b := reflect.ValueOf(e).Elem().FieldByName("Base").Interface().(*Base)

	h.call(e, b, t)
}

func (h *EventHandler) call(e interface{}, b *Base, t eventType) {
	filtered := h.applyMiddlewares(e, b, t)

	switch e := e.(type) {
	case *ReadyEvent:
		h.callHandlers(e, b, t, filtered, false)
		h.handleReady(e, filtered)
	case *GuildCreateEvent:
		h.callHandlers(e, b, t, filtered, true)
		h.handleGuildCreate(e, filtered)
	case *GuildDeleteEvent:
		h.callHandlers(e, b, t, filtered, true)
		h.handleGuildDelete(e, filtered)
	default:
		h.callHandlers(e, b, t, filtered, false)
	}
}

func (h *EventHandler) applyMiddlewares(e interface{}, b *Base, t eventType) (filtered bool) {
	h.middlewaresMutex.RLock()

	var wg sync.WaitGroup
	var aFiltered uint32

	wg.Add(len(h.middlewares[eventTypeAll]) +
		len(h.middlewares[eventTypeBase]) +
		len(h.middlewares[t]))

	h.startMiddlewares(e, eventTypeAll, &wg, &aFiltered)
	h.startMiddlewares(b, eventTypeBase, &wg, &aFiltered)
	h.startMiddlewares(e, t, &wg, &aFiltered)

	h.middlewaresMutex.RUnlock()

	wg.Wait()

	filtered = aFiltered == 1 // goroutines finished, no need for atomic ops anymore

	return
}

func (h *EventHandler) callHandlers(e interface{}, b *Base, t eventType, filtered, direct bool) {
	h.handlerMutex.RLock()

	var wg sync.WaitGroup

	if !direct {
		h.startHandlers(e, eventTypeAll, filtered, &wg)
		h.startHandlers(b, eventTypeBase, filtered, &wg)
	}

	h.startHandlers(e, t, filtered, &wg)

	h.handlerMutex.RUnlock()

	wg.Wait()
}

func (h *EventHandler) startHandlers(e interface{}, t eventType, filtered bool, wg *sync.WaitGroup) {
	for _, handler := range h.handler[t] {
		if !filtered || !handler.filtered {
			wg.Add(1)

			go func() {
				defer wg.Done()
				defer func() {
					if rec := recover(); rec != nil {
						h.PanicHandler(rec)
					}
				}()

				if err := handler.handler.handle(h.s, e); err != nil {
					h.ErrorHandler(err)
				}
			}()
		}
	}
}

func (h *EventHandler) startMiddlewares(e interface{}, t eventType, wg *sync.WaitGroup, aFiltered *uint32) {
	for _, mw := range h.middlewares[t] {
		go func(handler handler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					h.PanicHandler(r)
				}
			}()

			err := handler.handle(h.s, e)
			if err == Filtered {
				atomic.StoreUint32(aFiltered, 1)
			} else {
				h.ErrorHandler(err)
				return
			}
		}(mw)
	}
}

func (h *EventHandler) genEvent(src interface{}) (interface{}, *Base, eventType) {
	base := NewBase()

	switch src := src.(type) {
	// ---------------- Ready Event ----------------
	case *gateway.ReadyEvent:
		return &ReadyEvent{
			ReadyEvent: src,
			Base:       base,
		}, base, eventTypeReady

	// ---------------- Channel Events ----------------
	case *gateway.ChannelCreateEvent:
		return &ChannelCreateEvent{
			ChannelCreateEvent: src,
			Base:               base,
		}, base, eventTypeChannelCreate
	case *gateway.ChannelUpdateEvent:
		c, _ := h.s.Store.Channel(src.ID)

		return &ChannelUpdateEvent{
			ChannelUpdateEvent: src,
			Base:               base,
			Old:                c,
		}, base, eventTypeChannelUpdate
	case *gateway.ChannelDeleteEvent:
		c, _ := h.s.Store.Channel(src.ID)

		return &ChannelDeleteEvent{
			ChannelDeleteEvent: src,
			Base:               base,
			Old:                c,
		}, base, eventTypeChannelDelete
	case *gateway.ChannelPinsUpdateEvent:
		return &ChannelPinsUpdateEvent{
			ChannelPinsUpdateEvent: src,
			Base:                   base,
		}, base, eventTypeChannelPinsUpdate
	case *gateway.ChannelUnreadUpdateEvent:
		return &ChannelUnreadUpdateEvent{
			ChannelUnreadUpdateEvent: src,
			Base:                     base,
		}, base, eventTypeChannelUnreadUpdate

	// ---------------- Guild Events ----------------
	case *gateway.GuildCreateEvent:
		return &GuildCreateEvent{
			GuildCreateEvent: src,
			Base:             base,
		}, base, eventTypeGuildCreate
	case *gateway.GuildUpdateEvent:
		g, _ := h.s.Store.Guild(src.ID)

		return &GuildUpdateEvent{
			GuildUpdateEvent: src,
			Base:             base,
			Old:              g,
		}, base, eventTypeGuildUpdate
	case *gateway.GuildDeleteEvent:
		g, _ := h.s.Store.Guild(src.ID)

		return &GuildDeleteEvent{
			GuildDeleteEvent: src,
			Base:             base,
			Old:              g,
		}, base, eventTypeGuildDelete
	case *gateway.GuildBanAddEvent:
		return &GuildBanAddEvent{
			GuildBanAddEvent: src,
			Base:             base,
		}, base, eventTypeGuildBanAdd
	case *gateway.GuildBanRemoveEvent:
		return &GuildBanRemoveEvent{
			GuildBanRemoveEvent: src,
			Base:                base,
		}, base, eventTypeGuildBanRemove
	case *gateway.GuildEmojisUpdateEvent:
		e, _ := h.s.Store.Emojis(src.GuildID)

		return &GuildEmojisUpdateEvent{
			GuildEmojisUpdateEvent: src,
			Base:                   base,
			Old:                    e,
		}, base, eventTypeGuildEmojisUpdate
	case *gateway.GuildIntegrationsUpdateEvent:
		return &GuildIntegrationsUpdateEvent{
			GuildIntegrationsUpdateEvent: src,
			Base:                         base,
		}, base, eventTypeGuildIntegrationsUpdate
	case *gateway.GuildMemberAddEvent:
		return &GuildMemberAddEvent{
			GuildMemberAddEvent: src,
			Base:                base,
		}, base, eventTypeGuildMemberAdd
	case *gateway.GuildMemberRemoveEvent:
		m, _ := h.s.Store.Member(src.GuildID, src.User.ID)

		return &GuildMemberRemoveEvent{
			GuildMemberRemoveEvent: src,
			Base:                   base,
			Old:                    m,
		}, base, eventTypeGuildMemberRemove
	case *gateway.GuildMemberUpdateEvent:
		m, _ := h.s.Store.Member(src.GuildID, src.User.ID)

		return &GuildMemberUpdateEvent{
			GuildMemberUpdateEvent: src,
			Base:                   base,
			Old:                    m,
		}, base, eventTypeGuildMemberUpdate
	case *gateway.GuildMembersChunkEvent:
		return &GuildMembersChunkEvent{
			GuildMembersChunkEvent: src,
			Base:                   base,
		}, base, eventTypeGuildMembersChunk
	case *gateway.GuildRoleCreateEvent:
		return &GuildRoleCreateEvent{
			GuildRoleCreateEvent: src,
			Base:                 base,
		}, base, eventTypeGuildRoleCreate
	case *gateway.GuildRoleUpdateEvent:
		r, _ := h.s.Store.Role(src.GuildID, src.Role.ID)

		return &GuildRoleUpdateEvent{
			GuildRoleUpdateEvent: src,
			Base:                 base,
			Old:                  r,
		}, base, eventTypeGuildRoleUpdate
	case *gateway.GuildRoleDeleteEvent:
		r, _ := h.s.Store.Role(src.GuildID, src.RoleID)

		return &GuildRoleDeleteEvent{
			GuildRoleDeleteEvent: src,
			Base:                 base,
			Old:                  r,
		}, base, eventTypeGuildRoleDelete

	// ---------------- Invite Events ----------------
	case *gateway.InviteCreateEvent:
		return &InviteCreateEvent{
			InviteCreateEvent: src,
			Base:              base,
		}, base, eventTypeInviteCreate
	case *gateway.InviteDeleteEvent:
		return &InviteDeleteEvent{
			InviteDeleteEvent: src,
			Base:              base,
		}, base, eventTypeInviteDelete

	// ---------------- Message Events ----------------
	case *gateway.MessageCreateEvent:
		return &MessageCreateEvent{
			MessageCreateEvent: src,
			Base:               base,
		}, base, eventTypeMessageCreate
	case *gateway.MessageUpdateEvent:
		m, _ := h.s.Store.Message(src.ChannelID, src.ID)

		return &MessageUpdateEvent{
			MessageUpdateEvent: src,
			Base:               base,
			Old:                m,
		}, base, eventTypeMessageUpdate
	case *gateway.MessageDeleteEvent:
		m, _ := h.s.Store.Message(src.ChannelID, src.ID)

		return &MessageDeleteEvent{
			MessageDeleteEvent: src,
			Base:               base,
			Old:                m,
		}, base, eventTypeMessageDelete
	case *gateway.MessageDeleteBulkEvent:
		msgs := make([]discord.Message, 0, len(src.IDs))

		for _, id := range src.IDs {
			m, err := h.s.Store.Message(src.ChannelID, id)
			if err == nil {
				msgs = append(msgs, *m)
			}
		}

		return &MessageDeleteBulkEvent{
			MessageDeleteBulkEvent: src,
			Base:                   base,
			Old:                    msgs,
		}, base, eventTypeMessageDeleteBulk
	case *gateway.MessageReactionAddEvent:
		return &MessageReactionAddEvent{
			MessageReactionAddEvent: src,
			Base:                    base,
		}, base, eventTypeMessageReactionAdd
	case *gateway.MessageReactionRemoveEvent:
		return &MessageReactionRemoveEvent{
			MessageReactionRemoveEvent: src,
			Base:                       base,
		}, base, eventTypeMessageReactionRemove
	case *gateway.MessageReactionRemoveAllEvent:
		return &MessageReactionRemoveAllEvent{
			MessageReactionRemoveAllEvent: src,
			Base:                          base,
		}, base, eventTypeMessageReactionRemoveAll
	case *gateway.MessageReactionRemoveEmoji:
		return &MessageReactionRemoveEmojiEvent{
			MessageReactionRemoveEmoji: src,
			Base:                       base,
		}, base, eventTypeMessageReactionRemoveEmoji
	case *gateway.MessageAckEvent:
		return &MessageAckEvent{
			MessageAckEvent: src,
			Base:            base,
		}, base, eventTypeMessageAck

	// ---------------- Presence Events ----------------
	case *gateway.PresenceUpdateEvent:
		p, _ := h.s.Store.Presence(src.GuildID, src.User.ID)

		return &PresenceUpdateEvent{
			PresenceUpdateEvent: src,
			Base:                base,
			Old:                 p,
		}, base, eventTypePresenceUpdate
	case *gateway.PresencesReplaceEvent:
		return &PresencesReplaceEvent{
			PresencesReplaceEvent: src,
			Base:                  base,
		}, base, eventTypePresencesReplace
	case *gateway.SessionsReplaceEvent:
		return &SessionsReplaceEvent{
			SessionsReplaceEvent: src,
			Base:                 base,
		}, base, eventTypeSessionsReplace
	case *gateway.TypingStartEvent:
		return &TypingStartEvent{
			TypingStartEvent: src,
			Base:             base,
		}, base, eventTypeTypingStart
	case *gateway.UserUpdateEvent:
		return &UserUpdateEvent{
			UserUpdateEvent: src,
			Base:            base,
		}, base, eventTypeUserUpdate

	// ---------------- Relationship Events ----------------
	case *gateway.RelationshipAddEvent:
		return &RelationshipAddEvent{
			RelationshipAddEvent: src,
			Base:                 base,
		}, base, eventTypeRelationshipAdd
	case *gateway.RelationshipRemoveEvent:
		return &RelationshipRemoveEvent{
			RelationshipRemoveEvent: src,
			Base:                    base,
		}, base, eventTypeRelationshipRemove

	// ---------------- User Settings Events ----------------
	case *gateway.UserGuildSettingsUpdateEvent:
		return &UserGuildSettingsUpdateEvent{
			UserGuildSettingsUpdateEvent: src,
			Base:                         base,
		}, base, eventTypeUserGuildSettingsUpdate
	case *gateway.UserSettingsUpdateEvent:
		return &UserSettingsUpdateEvent{
			UserSettingsUpdateEvent: src,
			Base:                    base,
		}, base, eventTypeUserSettingsUpdate
	case *gateway.UserNoteUpdateEvent:
		return &UserNoteUpdateEvent{
			UserNoteUpdateEvent: src,
			Base:                base,
		}, base, eventTypeUserNoteUpdate

	// ---------------- Voice Events ----------------
	case *gateway.VoiceStateUpdateEvent:
		return &VoiceStateUpdateEvent{
			VoiceStateUpdateEvent: src,
			Base:                  base,
		}, base, eventTypeVoiceStateUpdate
	case *gateway.VoiceServerUpdateEvent:
		return &VoiceServerUpdateEvent{
			VoiceServerUpdateEvent: src,
			Base:                   base,
		}, base, eventTypeVoiceServerUpdate

	// ---------------- Webhook Events ----------------
	case *gateway.WebhooksUpdateEvent:
		return &WebhooksUpdateEvent{
			WebhooksUpdateEvent: src,
			Base:                base,
		}, base, eventTypeWebhooksUpdate
	}

	return nil, nil, 0
}

func (h *EventHandler) handleReady(e *ReadyEvent, filtered bool) {
	for _, g := range e.Guilds {
		// store this so we know when we need to dispatch a belated
		// GuildReadyEvent
		if g.Unavailable {
			h.s.unreadyGuilds.Add(g.ID)
		} else {
			h.callHandlers(&GuildReadyEvent{
				GuildCreateEvent: &GuildCreateEvent{
					GuildCreateEvent: &g,
					Base:             NewBase(),
				},
			}, e.Base, eventTypeGuildReady, filtered, false)
		}
	}
}

func (h *EventHandler) handleGuildCreate(e *GuildCreateEvent, filtered bool) {
	// this guild was unavailable, but has come back online
	if h.s.unavailableGuilds.Delete(e.ID) {
		h.callHandlers(&GuildAvailableEvent{
			GuildCreateEvent: e,
		}, e.Base, eventTypeGuildAvailable, filtered, false)

		// the guild was already unavailable when connecting to the gateway
		// we can dispatch a belated GuildReadyEvent
	} else if h.s.unreadyGuilds.Delete(e.ID) {
		h.callHandlers(&GuildReadyEvent{
			GuildCreateEvent: e,
		}, e.Base, eventTypeGuildReady, filtered, false)
	} else { // we don't know this guild, hence we just joined it
		h.callHandlers(&GuildJoinEvent{
			GuildCreateEvent: e,
		}, e.Base, eventTypeGuildJoin, filtered, false)
	}
}

func (h *EventHandler) handleGuildDelete(e *GuildDeleteEvent, filtered bool) {
	// store this so we can later dispatch a GuildAvailableEvent, once the
	// guild becomes available again.
	if e.Unavailable {
		h.s.unavailableGuilds.Add(e.ID)

		h.callHandlers(&GuildUnavailableEvent{
			GuildDeleteEvent: e,
		}, e.Base, eventTypeGuildUnavailable, filtered, false)
	} else {
		// it might have been unavailable before we left
		h.s.unavailableGuilds.Delete(e.ID)

		h.callHandlers(&GuildLeaveEvent{
			GuildDeleteEvent: e,
		}, e.Base, eventTypeGuildLeave, filtered, false)
	}
}
