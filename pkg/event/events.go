package event

// Code generated by tools/codegen/event. DO NOT EDIT.

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type (
	InvalidSession struct {
		*Base
		*gateway.InvalidSessionEvent
	}
	GuildMemberUpdate struct {
		*Base
		*gateway.GuildMemberUpdateEvent

		Old *discord.Member
	}
	GuildMembersChunk struct {
		*Base
		*gateway.GuildMembersChunkEvent
	}
	PresenceUpdate struct {
		*Base
		*gateway.PresenceUpdateEvent

		Old *gateway.Presence
	}
	WebhooksUpdate struct {
		*Base
		*gateway.WebhooksUpdateEvent
	}
	ApplicationCommandUpdate struct {
		*Base
		*gateway.ApplicationCommandUpdateEvent
	}
	ChannelDelete struct {
		*Base
		*gateway.ChannelDeleteEvent

		Old *discord.Channel
	}
	GuildUpdate struct {
		*Base
		*gateway.GuildUpdateEvent

		Old *discord.Guild
	}
	GuildBanRemove struct {
		*Base
		*gateway.GuildBanRemoveEvent
	}
	MessageUpdate struct {
		*Base
		*gateway.MessageUpdateEvent

		Old *discord.Message
	}
	MessageDelete struct {
		*Base
		*gateway.MessageDeleteEvent

		Old *discord.Message
	}
	MessageReactionRemove struct {
		*Base
		*gateway.MessageReactionRemoveEvent
	}
	RelationshipRemove struct {
		*Base
		*gateway.RelationshipRemoveEvent
	}
	GuildCreate struct {
		*Base
		*gateway.GuildCreateEvent
	}
	GuildDelete struct {
		*Base
		*gateway.GuildDeleteEvent

		Old *discord.Guild
	}
	GuildIntegrationsUpdate struct {
		*Base
		*gateway.GuildIntegrationsUpdateEvent
	}
	GuildMemberAdd struct {
		*Base
		*gateway.GuildMemberAddEvent
	}
	SessionsReplace struct {
		*Base
		*gateway.SessionsReplaceEvent
	}
	TypingStart struct {
		*Base
		*gateway.TypingStartEvent
	}
	ChannelUnreadUpdate struct {
		*Base
		*gateway.ChannelUnreadUpdateEvent
	}
	MessageAck struct {
		*Base
		*gateway.MessageAckEvent
	}
	UserSettingsUpdate struct {
		*Base
		*gateway.UserSettingsUpdateEvent
	}
	UserNoteUpdate struct {
		*Base
		*gateway.UserNoteUpdateEvent
	}
	PresencesReplace struct {
		*Base
		*gateway.PresencesReplaceEvent
	}
	Ready struct {
		*Base
		*gateway.ReadyEvent
	}
	ChannelCreate struct {
		*Base
		*gateway.ChannelCreateEvent
	}
	ChannelPinsUpdate struct {
		*Base
		*gateway.ChannelPinsUpdateEvent
	}
	GuildRoleCreate struct {
		*Base
		*gateway.GuildRoleCreateEvent
	}
	GuildRoleDelete struct {
		*Base
		*gateway.GuildRoleDeleteEvent

		Old *discord.Role
	}
	MessageDeleteBulk struct {
		*Base
		*gateway.MessageDeleteBulkEvent
	}
	MessageReactionRemoveEmoji struct {
		*Base
		*gateway.MessageReactionRemoveEmojiEvent
	}
	RelationshipAdd struct {
		*Base
		*gateway.RelationshipAddEvent
	}
	Hello struct {
		*Base
		*gateway.HelloEvent
	}
	ReadySupplemental struct {
		*Base
		*gateway.ReadySupplementalEvent
	}
	ChannelUpdate struct {
		*Base
		*gateway.ChannelUpdateEvent

		Old *discord.Channel
	}
	GuildEmojisUpdate struct {
		*Base
		*gateway.GuildEmojisUpdateEvent
	}
	GuildMemberListUpdate struct {
		*Base
		*gateway.GuildMemberListUpdate
	}
	MessageReactionAdd struct {
		*Base
		*gateway.MessageReactionAddEvent
	}
	Resumed struct {
		*Base
		*gateway.ResumedEvent
	}
	GuildBanAdd struct {
		*Base
		*gateway.GuildBanAddEvent
	}
	GuildRoleUpdate struct {
		*Base
		*gateway.GuildRoleUpdateEvent

		Old *discord.Role
	}
	MessageReactionRemoveAll struct {
		*Base
		*gateway.MessageReactionRemoveAllEvent
	}
	VoiceStateUpdate struct {
		*Base
		*gateway.VoiceStateUpdateEvent
	}
	UserGuildSettingsUpdate struct {
		*Base
		*gateway.UserGuildSettingsUpdateEvent
	}
	UserUpdate struct {
		*Base
		*gateway.UserUpdateEvent
	}
	GuildMemberRemove struct {
		*Base
		*gateway.GuildMemberRemoveEvent

		Old *discord.Member
	}
	InviteCreate struct {
		*Base
		*gateway.InviteCreateEvent
	}
	InviteDelete struct {
		*Base
		*gateway.InviteDeleteEvent
	}
	MessageCreate struct {
		*Base
		*gateway.MessageCreateEvent
	}
	VoiceServerUpdate struct {
		*Base
		*gateway.VoiceServerUpdateEvent
	}
	InteractionCreate struct {
		*Base
		*gateway.InteractionCreateEvent
	}
)
