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

func calcEventType(e interface{}) eventType {
	switch e.(type) {
	// ---------------- Ready Event ----------------
	case *ReadyEvent:
		return eventTypeReady

	// ---------------- Channel Events ----------------
	case *ChannelCreateEvent:
		return eventTypeChannelCreate
	case *ChannelUpdateEvent:
		return eventTypeChannelUpdate
	case *ChannelDeleteEvent:
		return eventTypeChannelDelete
	case *ChannelPinsUpdateEvent:
		return eventTypeChannelPinsUpdate
	case *ChannelUnreadUpdateEvent:
		return eventTypeChannelUnreadUpdate

	// ---------------- Guild Events ----------------
	// -------- Guild Create Events --------
	case *GuildCreateEvent:
		return eventTypeGuildCreate
	case *GuildReadyEvent:
		return eventTypeGuildReady
	case *GuildAvailableEvent:
		return eventTypeGuildAvailable
	case *GuildJoinEvent:
		return eventTypeGuildJoin

	case *GuildUpdateEvent:
		return eventTypeGuildUpdate

	// -------- Guild Delete Events --------
	case *GuildDeleteEvent:
		return eventTypeGuildDelete
	case *GuildLeaveEvent:
		return eventTypeGuildLeave
	case *GuildUnavailableEvent:
		return eventTypeGuildUnavailable

	case *GuildBanAddEvent:
		return eventTypeGuildBanAdd
	case *GuildBanRemoveEvent:
		return eventTypeGuildBanRemove
	case *GuildEmojisUpdateEvent:
		return eventTypeGuildEmojisUpdate
	case *GuildIntegrationsUpdateEvent:
		return eventTypeGuildIntegrationsUpdate
	case *GuildMemberAddEvent:
		return eventTypeGuildMemberAdd
	case *GuildMemberRemoveEvent:
		return eventTypeGuildMemberRemove
	case *GuildMemberUpdateEvent:
		return eventTypeGuildMemberUpdate
	case *GuildMembersChunkEvent:
		return eventTypeGuildMembersChunk
	case *GuildRoleCreateEvent:
		return eventTypeGuildRoleCreate
	case *GuildRoleUpdateEvent:
		return eventTypeGuildRoleUpdate
	case *GuildRoleDeleteEvent:
		return eventTypeGuildRoleDelete

	// ---------------- Invite Events ----------------
	case *InviteCreateEvent:
		return eventTypeInviteCreate
	case *InviteDeleteEvent:
		return eventTypeInviteDelete

	// ---------------- Message Events ----------------
	case *MessageCreateEvent:
		return eventTypeMessageCreate
	case *MessageUpdateEvent:
		return eventTypeMessageUpdate
	case *MessageDeleteEvent:
		return eventTypeMessageDelete
	case *MessageDeleteBulkEvent:
		return eventTypeMessageDeleteBulk
	case *MessageReactionAddEvent:
		return eventTypeMessageReactionAdd
	case *MessageReactionRemoveEvent:
		return eventTypeMessageReactionRemove
	case *MessageReactionRemoveAllEvent:
		return eventTypeMessageReactionRemoveAll
	case *MessageReactionRemoveEmojiEvent:
		return eventTypeMessageReactionRemoveEmoji
	case *MessageAckEvent:
		return eventTypeMessageAck

	// ---------------- Presence Events ----------------
	case *PresenceUpdateEvent:
		return eventTypePresenceUpdate
	case *PresencesReplaceEvent:
		return eventTypePresencesReplace
	case *SessionsReplaceEvent:
		return eventTypeSessionsReplace
	case *TypingStartEvent:
		return eventTypeTypingStart
	case *UserUpdateEvent:
		return eventTypeUserUpdate

	// ---------------- Relationship Events ----------------
	case *RelationshipAddEvent:
		return eventTypeRelationshipAdd
	case *RelationshipRemoveEvent:
		return eventTypeRelationshipRemove

	// ---------------- User Settings Events ----------------
	case *UserGuildSettingsUpdateEvent:
		return eventTypeUserGuildSettingsUpdate
	case *UserSettingsUpdateEvent:
		return eventTypeUserSettingsUpdate
	case *UserNoteUpdateEvent:
		return eventTypeUserNoteUpdate

	// ---------------- Voice Events ----------------
	case *VoiceStateUpdateEvent:
		return eventTypeVoiceStateUpdate
	case *VoiceServerUpdateEvent:
		return eventTypeVoiceServerUpdate

	// ---------------- Webhook Events ----------------
	case *WebhooksUpdateEvent:
		return eventTypeWebhooksUpdate

	// ---------------- Webhook Events ----------------
	case *CloseEvent:
		return eventTypeClose
	}

	return 0
}
