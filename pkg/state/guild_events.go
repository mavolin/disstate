package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
)

// ================================ Guild Create ================================

// https://discord.com/developers/docs/topics/gateway#guild-create
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildCreateEvent struct {
	*gateway.GuildCreateEvent
	*Base
}

type guildCreateEventHandler func(s *State, e *GuildCreateEvent) error

func (h guildCreateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildCreateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Ready ================================

// GuildReadyEvent is a situation-specific GuildCreate event.
// It gets fired during Ready for all available guilds.
// Additionally, it gets fired for all those guilds that become available after
// initially connecting, but were not during Ready.
type GuildReadyEvent struct {
	*GuildCreateEvent
}

type guildReadyEventHandler func(s *State, e *GuildReadyEvent) error

func (h guildReadyEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildReadyEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Available ================================

// GuildAvailableEvent is a situation-specific GuildCreate event.
// It gets fired when a guild becomes available, after getting marked
// unavailable during a GuildUnavailableEvent event.
// This event will not be fired for guilds that were already unavailable when
// initially connecting.
type GuildAvailableEvent struct {
	*GuildCreateEvent
}

type guildAvailableEventHandler func(s *State, e *GuildAvailableEvent) error

func (h guildAvailableEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildAvailableEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Join ================================

// GuildJoinEvent is a situation-specific GuildCreate event.
// It gets fired when the user/bot joins a guild.
type GuildJoinEvent struct {
	*GuildCreateEvent
}

type guildJoinEventHandler func(s *State, e *GuildJoinEvent) error

func (h guildJoinEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildJoinEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Update ================================

// https://discord.com/developers/docs/topics/gateway#guild-update
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildUpdateEvent struct {
	*gateway.GuildUpdateEvent
	*Base

	Old *discord.Guild
}

type guildUpdateEventHandler func(s *State, e *GuildUpdateEvent) error

func (h guildUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Delete ================================

// https://discord.com/developers/docs/topics/gateway#guild-delete
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildDeleteEvent struct {
	*gateway.GuildDeleteEvent
	*Base

	Old *discord.Guild
}

type guildDeleteEventHandler func(s *State, e *GuildDeleteEvent) error

func (h guildDeleteEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildDeleteEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Leave ================================

// GuildLeaveEvent is a situation-specific GuildDeleteEvent event.
// It gets fired when the user/bot leaves guild, gets kicked/banned from it, or
// the owner deletes it.
type GuildLeaveEvent struct {
	*GuildDeleteEvent
}

type guildLeaveEventHandler func(s *State, e *GuildLeaveEvent) error

func (h guildLeaveEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildLeaveEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Unavailable ================================

// GuildUnavailableEvent is a situation-specific GuildDeleteEvent event.
// It gets fired if the guild becomes unavailable, e.g. through a discord
// outage.
type GuildUnavailableEvent struct {
	*GuildDeleteEvent
}

type guildUnavailableEventHandler func(s *State, e *GuildUnavailableEvent) error

func (h guildUnavailableEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildUnavailableEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Ban Add ================================

// https://discord.com/developers/docs/topics/gateway#guild-ban-add
type GuildBanAddEvent struct {
	*gateway.GuildBanAddEvent
	*Base
}

type guildBanAddEventHandler func(s *State, e *GuildBanAddEvent) error

func (h guildBanAddEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildBanAddEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Ban Remove ================================

// https://discord.com/developers/docs/topics/gateway#guild-ban-remove
type GuildBanRemoveEvent struct {
	*gateway.GuildBanRemoveEvent
	*Base
}

type guildBanRemoveEventHandler func(s *State, e *GuildBanRemoveEvent) error

func (h guildBanRemoveEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildBanRemoveEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Emojis Update ================================

// https://discord.com/developers/docs/topics/gateway#guild-emojis-update
type GuildEmojisUpdateEvent struct {
	*gateway.GuildEmojisUpdateEvent
	*Base

	Old []discord.Emoji
}

type guildEmojisUpdateEventHandler func(s *State, e *GuildEmojisUpdateEvent) error

func (h guildEmojisUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildEmojisUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Integrations Update ================================

// https://discord.com/developers/docs/topics/gateway#guild-integrations-update
type GuildIntegrationsUpdateEvent struct {
	*gateway.GuildIntegrationsUpdateEvent
	*Base
}

type guildIntegrationsUpdateEventHandler func(s *State, e *GuildIntegrationsUpdateEvent) error

func (h guildIntegrationsUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildIntegrationsUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Member Add ================================

// https://discord.com/developers/docs/topics/gateway#guild-member-add
type GuildMemberAddEvent struct {
	*gateway.GuildMemberAddEvent
	*Base
}

type guildMemberAddEventHandler func(s *State, e *GuildMemberAddEvent) error

func (h guildMemberAddEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildMemberAddEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Member Remove ================================

// https://discord.com/developers/docs/topics/gateway#guild-member-remove
type GuildMemberRemoveEvent struct {
	*gateway.GuildMemberRemoveEvent
	*Base

	Old *discord.Member
}

type guildMemberRemoveEventHandler func(s *State, e *GuildMemberRemoveEvent) error

func (h guildMemberRemoveEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildMemberRemoveEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Member Update ================================

// https://discord.com/developers/docs/topics/gateway#guild-member-update
type GuildMemberUpdateEvent struct {
	*gateway.GuildMemberUpdateEvent
	*Base

	Old *discord.Member
}

type guildMemberUpdateEventHandler func(s *State, e *GuildMemberUpdateEvent) error

func (h guildMemberUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildMemberUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Members Chunk ================================

// https://discord.com/developers/docs/topics/gateway#guild-members-chunk
type GuildMembersChunkEvent struct {
	*gateway.GuildMembersChunkEvent
	*Base
}

type guildMembersChunkEventHandler func(s *State, e *GuildMembersChunkEvent) error

func (h guildMembersChunkEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildMembersChunkEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Role Create ================================

// https://discord.com/developers/docs/topics/gateway#guild-role-create
type GuildRoleCreateEvent struct {
	*gateway.GuildRoleCreateEvent
	*Base
}

type guildRoleCreateEventHandler func(s *State, e *GuildRoleCreateEvent) error

func (h guildRoleCreateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildRoleCreateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Role Update ================================

// https://discord.com/developers/docs/topics/gateway#guild-role-update
type GuildRoleUpdateEvent struct {
	*gateway.GuildRoleUpdateEvent
	*Base

	Old *discord.Role
}

type guildRoleUpdateEventHandler func(s *State, e *GuildRoleUpdateEvent) error

func (h guildRoleUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildRoleUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}

// ================================ Guild Role Delete ================================

// https://discord.com/developers/docs/topics/gateway#guild-role-delete
type GuildRoleDeleteEvent struct {
	*gateway.GuildRoleDeleteEvent
	*Base

	Old *discord.Role
}

type guildRoleDeleteEventHandler func(s *State, e *GuildRoleDeleteEvent) error

func (h guildRoleDeleteEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*GuildRoleDeleteEvent); ok {
		return h(s, e)
	}

	return nil
}
