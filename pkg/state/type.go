package state

type eventType uint8

const (
	// ---- Generic Events ----
	eventTypeAll eventType = iota // interface{}
	eventTypeBase

	// ---- Ready GuildRoleDeleteEvent ----
	eventTypeReady

	// ---- Channel Events ----
	eventTypeChannelCreate
	eventTypeChannelUpdate
	eventTypeChannelDelete
	eventTypeChannelPinsUpdate
	eventTypeChannelUnreadUpdate

	// ---- Guild Events ----
	// -- Guild Create Events --
	eventTypeGuildCreate
	eventTypeGuildReady
	eventTypeGuildAvailable
	eventTypeGuildJoin
	// ----
	eventTypeGuildUpdate
	// -- Guild Delete Events --
	eventTypeGuildDelete
	eventTypeGuildLeave
	eventTypeGuildUnavailable
	// ----
	eventTypeGuildBanAdd
	eventTypeGuildBanRemove
	eventTypeGuildEmojisUpdate
	eventTypeGuildIntegrationsUpdate
	eventTypeGuildMemberAdd
	eventTypeGuildMemberRemove
	eventTypeGuildMemberUpdate
	eventTypeGuildMembersChunk
	eventTypeGuildRoleCreate
	eventTypeGuildRoleUpdate
	eventTypeGuildRoleDelete

	// ---- Invite Events ----
	eventTypeInviteCreate
	eventTypeInviteDelete

	// ---- Message Events ----
	eventTypeMessageCreate
	eventTypeMessageUpdate
	eventTypeMessageDelete
	eventTypeMessageDeleteBulk
	eventTypeMessageReactionAdd
	eventTypeMessageReactionRemove
	eventTypeMessageReactionRemoveAll
	eventTypeMessageReactionRemoveEmoji
	eventTypeMessageAck

	// ---- Presence Events ----
	eventTypePresenceUpdate
	eventTypePresencesReplace
	eventTypeSessionsReplace
	eventTypeTypingStart
	eventTypeUserUpdate

	// ---- Relationship Events ----
	eventTypeRelationshipAdd
	eventTypeRelationshipRemove

	// ---- User Settings Events ----
	eventTypeUserGuildSettingsUpdate
	eventTypeUserSettingsUpdate
	eventTypeUserNoteUpdate

	// ---- Voice Events ----
	eventTypeVoiceStateUpdate
	eventTypeVoiceServerUpdate

	// ---- Webhook Events ----
	eventTypeWebhooksUpdate

	// ---- Custom Events ----
	eventTypeClose
)
