package state

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

// https://discord.com/developers/docs/topics/gateway#guild-create
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildCreateEvent struct {
	*gateway.GuildCreateEvent
	*Base
}

// GuildReadyEvent is a situation-specific GuildCreate event.
// It gets fired during Ready for all available guilds.
// Additionally, it gets fired for all those guilds that become available after
// initially connecting, but were not during Ready.
type GuildReadyEvent struct {
	*GuildCreateEvent
}

// GuildAvailableEvent is a situation-specific GuildCreate event.
// It gets fired when a guild becomes available, after getting marked
// unavailable during a GuildUnavailableEvent event.
// This event will not be fired for guilds that were already unavailable when
// initially connecting.
type GuildAvailableEvent struct {
	*GuildCreateEvent
}

// GuildJoinEvent is a situation-specific GuildCreate event.
// It gets fired when the user/bot joins a guild.
type GuildJoinEvent struct {
	*GuildCreateEvent
}

// https://discord.com/developers/docs/topics/gateway#guild-update
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildUpdateEvent struct {
	*gateway.GuildUpdateEvent
	*Base

	Old *discord.Guild
}

// https://discord.com/developers/docs/topics/gateway#guild-delete
//
// Note that this event will not be sent in Base and All handlers.
// Instead, the situation-specific sub-events will be sent.
type GuildDeleteEvent struct {
	*gateway.GuildDeleteEvent
	*Base

	Old *discord.Guild
}

// GuildLeaveEvent is a situation-specific GuildDeleteEvent event.
// It gets fired when the user/bot leaves guild, gets kicked/banned from it, or
// the owner deletes it.
type GuildLeaveEvent struct {
	*GuildDeleteEvent
}

// GuildUnavailableEvent is a situation-specific GuildDeleteEvent event.
// It gets fired if the guild becomes unavailable, e.g. through a discord
// outage.
type GuildUnavailableEvent struct {
	*GuildDeleteEvent
}

// https://discord.com/developers/docs/topics/gateway#guild-ban-add
type GuildBanAddEvent struct {
	*gateway.GuildBanAddEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-ban-remove
type GuildBanRemoveEvent struct {
	*gateway.GuildBanRemoveEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-emojis-update
type GuildEmojisUpdateEvent struct {
	*gateway.GuildEmojisUpdateEvent
	*Base

	Old []discord.Emoji
}

// https://discord.com/developers/docs/topics/gateway#guild-integrations-update
type GuildIntegrationsUpdateEvent struct {
	*gateway.GuildIntegrationsUpdateEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-member-add
type GuildMemberAddEvent struct {
	*gateway.GuildMemberAddEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-member-remove
type GuildMemberRemoveEvent struct {
	*gateway.GuildMemberRemoveEvent
	*Base

	Old *discord.Member
}

// https://discord.com/developers/docs/topics/gateway#guild-member-update
type GuildMemberUpdateEvent struct {
	*gateway.GuildMemberUpdateEvent
	*Base

	Old *discord.Member
}

// https://discord.com/developers/docs/topics/gateway#guild-members-chunk
type GuildMembersChunkEvent struct {
	*gateway.GuildMembersChunkEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-role-create
type GuildRoleCreateEvent struct {
	*gateway.GuildRoleCreateEvent
	*Base
}

// https://discord.com/developers/docs/topics/gateway#guild-role-update
type GuildRoleUpdateEvent struct {
	*gateway.GuildRoleUpdateEvent
	*Base

	Old *discord.Role
}

// https://discord.com/developers/docs/topics/gateway#guild-role-delete
type GuildRoleDeleteEvent struct {
	*gateway.GuildRoleDeleteEvent
	*Base

	Old *discord.Role
}
