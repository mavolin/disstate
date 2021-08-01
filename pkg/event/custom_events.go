package event

// Close is a custom event that gets dispatched when the gateway closes.
type Close struct {
	*Base
}

// GuildReady is a situation-specific GuildCreate event.
// It gets fired during Ready for all available guilds.
// Additionally, it gets fired for all those guilds that become available after
// initially connecting, but were not during Ready.
type GuildReady struct {
	*GuildCreate
}

// GuildAvailable is a situation-specific GuildCreate event.
// It gets fired when a guild becomes available, after getting marked
// unavailable during a GuildUnavailableEvent event.
// This event will not be fired for guilds that were already unavailable when
// initially connecting.
type GuildAvailable struct {
	*GuildCreate
}

// GuildJoin is a situation-specific GuildCreate event.
// It gets fired when the user/bot joins a guild.
type GuildJoin struct {
	*GuildCreate
}

// GuildUnavailable is a situation-specific GuildDeleteEvent event.
// It gets fired if the guild becomes unavailable, e.g. through a discord
// outage.
type GuildUnavailable struct {
	*GuildDelete
}

// GuildLeave is a situation-specific GuildDeleteEvent event.
// It gets fired when the user/bot leaves guild, gets kicked/banned from it, or
// the owner deletes it.
type GuildLeave struct {
	*GuildDelete
}
