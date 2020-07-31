package state

type handler interface {
	handle(s *State, e interface{}) error
}

func handlerFuncForHandler(handlerFunc interface{}) (handler, eventType) {
	switch handlerFunc := handlerFunc.(type) {
	// ---------------- Generic Events ----------------
	case func(*State, interface{}) error:
		return allHandler(handlerFunc), eventTypeAll
	case func(*State, *Base) error:
		return baseHandler(handlerFunc), eventTypeBase

	// ---------------- Ready Event ----------------
	case func(*State, *ReadyEvent) error:
		return readyEventHandler(handlerFunc), eventTypeReady

	// ---------------- Channel Events ----------------
	case func(*State, *ChannelCreateEvent) error:
		return channelCreateEventHandler(handlerFunc), eventTypeChannelCreate
	case func(*State, *ChannelUpdateEvent) error:
		return channelUpdateEventHandler(handlerFunc), eventTypeChannelUpdate
	case func(*State, *ChannelDeleteEvent) error:
		return channelDeleteEventHandler(handlerFunc), eventTypeChannelDelete
	case func(*State, *ChannelPinsUpdateEvent) error:
		return channelPinsUpdateEventHandler(handlerFunc), eventTypeChannelPinsUpdate
	case func(*State, *ChannelUnreadUpdateEvent) error:
		return channelUnreadUpdateEventHandler(handlerFunc), eventTypeChannelUnreadUpdate

	// ---------------- Guild Events ----------------
	// -------- Guild Create Events --------
	case func(*State, *GuildCreateEvent) error:
		return guildCreateEventHandler(handlerFunc), eventTypeGuildCreate
	case func(*State, *GuildReadyEvent) error:
		return guildReadyEventHandler(handlerFunc), eventTypeGuildReady
	case func(*State, *GuildAvailableEvent) error:
		return guildAvailableEventHandler(handlerFunc), eventTypeGuildAvailable
	case func(*State, *GuildJoinEvent) error:
		return guildJoinEventHandler(handlerFunc), eventTypeGuildJoin

	case func(*State, *GuildUpdateEvent) error:
		return guildUpdateEventHandler(handlerFunc), eventTypeGuildUpdate

	// -------- Guild Delete Events --------
	case func(*State, *GuildDeleteEvent) error:
		return guildDeleteEventHandler(handlerFunc), eventTypeGuildDelete
	case func(*State, *GuildLeaveEvent) error:
		return guildLeaveEventHandler(handlerFunc), eventTypeGuildLeave
	case func(*State, *GuildUnavailableEvent) error:
		return guildUnavailableEventHandler(handlerFunc), eventTypeGuildUnavailable

	case func(*State, *GuildBanAddEvent) error:
		return guildBanAddEventHandler(handlerFunc), eventTypeGuildBanAdd
	case func(*State, *GuildBanRemoveEvent) error:
		return guildBanRemoveEventHandler(handlerFunc), eventTypeGuildBanRemove
	case func(*State, *GuildEmojisUpdateEvent) error:
		return guildEmojisUpdateEventHandler(handlerFunc), eventTypeGuildEmojisUpdate
	case func(*State, *GuildIntegrationsUpdateEvent) error:
		return guildIntegrationsUpdateEventHandler(handlerFunc), eventTypeGuildIntegrationsUpdate
	case func(*State, *GuildMemberAddEvent) error:
		return guildMemberAddEventHandler(handlerFunc), eventTypeGuildMemberAdd
	case func(*State, *GuildMemberRemoveEvent) error:
		return guildMemberRemoveEventHandler(handlerFunc), eventTypeGuildMemberRemove
	case func(*State, *GuildMemberUpdateEvent) error:
		return guildMemberUpdateEventHandler(handlerFunc), eventTypeGuildMemberUpdate
	case func(*State, *GuildMembersChunkEvent) error:
		return guildMembersChunkEventHandler(handlerFunc), eventTypeGuildMembersChunk
	case func(*State, *GuildRoleCreateEvent) error:
		return guildRoleCreateEventHandler(handlerFunc), eventTypeGuildRoleCreate
	case func(*State, *GuildRoleUpdateEvent) error:
		return guildRoleUpdateEventHandler(handlerFunc), eventTypeGuildRoleUpdate
	case func(*State, *GuildRoleDeleteEvent) error:
		return guildRoleDeleteEventHandler(handlerFunc), eventTypeGuildRoleDelete

	// ---------------- Invite Events ----------------
	case func(*State, *InviteCreateEvent) error:
		return inviteCreateEventHandler(handlerFunc), eventTypeInviteCreate
	case func(*State, *InviteDeleteEvent) error:
		return inviteDeleteEventHandler(handlerFunc), eventTypeInviteDelete

	// ---------------- Message Events ----------------
	case func(*State, *MessageCreateEvent) error:
		return messageCreateEventHandler(handlerFunc), eventTypeMessageCreate
	case func(*State, *MessageUpdateEvent) error:
		return messageUpdateEventHandler(handlerFunc), eventTypeMessageUpdate
	case func(*State, *MessageDeleteEvent) error:
		return messageDeleteEventHandler(handlerFunc), eventTypeMessageDelete
	case func(*State, *MessageDeleteBulkEvent) error:
		return messageDeleteBulkEventHandler(handlerFunc), eventTypeMessageDeleteBulk
	case func(*State, *MessageReactionAddEvent) error:
		return messageReactionAddEventHandler(handlerFunc), eventTypeMessageReactionAdd
	case func(*State, *MessageReactionRemoveEvent) error:
		return messageReactionRemoveEventHandler(handlerFunc), eventTypeMessageReactionRemove
	case func(*State, *MessageReactionRemoveAllEvent) error:
		return messageReactionRemoveAllEventHandler(handlerFunc), eventTypeMessageReactionRemoveAll
	case func(*State, *MessageReactionRemoveEmojiEvent) error:
		return messageReactionRemoveEmojiEventHandler(handlerFunc), eventTypeMessageReactionRemoveEmoji
	case func(*State, *MessageAckEvent) error:
		return messageAckEventHandler(handlerFunc), eventTypeMessageAck

	// ---------------- Presence Events ----------------
	case func(*State, *PresenceUpdateEvent) error:
		return presenceUpdateEventHandler(handlerFunc), eventTypePresenceUpdate
	case func(*State, *PresencesReplaceEvent) error:
		return presencesReplaceEventHandler(handlerFunc), eventTypePresencesReplace
	case func(*State, *SessionsReplaceEvent) error:
		return sessionsReplaceEventHandler(handlerFunc), eventTypeSessionsReplace
	case func(*State, *TypingStartEvent) error:
		return typingStartEventHandler(handlerFunc), eventTypeTypingStart
	case func(*State, *UserUpdateEvent) error:
		return userUpdateEventHandler(handlerFunc), eventTypeUserUpdate

	// ---------------- Relationship Events ----------------
	case func(*State, *RelationshipAddEvent) error:
		return relationshipAddEventHandler(handlerFunc), eventTypeRelationshipAdd
	case func(*State, *RelationshipRemoveEvent) error:
		return relationshipRemoveEventHandler(handlerFunc), eventTypeRelationshipRemove

	// ---------------- User Settings Events ----------------
	case func(*State, *UserGuildSettingsUpdateEvent) error:
		return userGuildSettingsUpdateEventHandler(handlerFunc), eventTypeUserGuildSettingsUpdate
	case func(*State, *UserSettingsUpdateEvent) error:
		return userSettingsUpdateEventHandler(handlerFunc), eventTypeUserSettingsUpdate
	case func(*State, *UserNoteUpdateEvent) error:
		return userNoteUpdateEventHandler(handlerFunc), eventTypeUserNoteUpdate

	// ---------------- Voice Events ----------------
	case func(*State, *VoiceStateUpdateEvent) error:
		return voiceStateUpdateEventHandler(handlerFunc), eventTypeVoiceStateUpdate
	case func(*State, *VoiceServerUpdateEvent) error:
		return voiceServerUpdateEventHandler(handlerFunc), eventTypeVoiceServerUpdate

	// ---------------- Webhook Events ----------------
	case func(*State, *WebhooksUpdateEvent) error:
		return webhooksUpdateEventHandler(handlerFunc), eventTypeWebhooksUpdate

	// ---------------- Webhook Events ----------------
	case func(*State, *CloseEvent) error:
		return closeEventHandler(handlerFunc), eventTypeClose
	}

	return nil, 0
}
