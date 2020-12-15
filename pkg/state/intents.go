package state

import (
	"reflect"

	"github.com/diamondburned/arikawa/v2/gateway"
)

var eventIntents = map[reflect.Type]gateway.Intents{
	reflect.TypeOf(new(GuildCreateEvent)):       gateway.IntentGuilds,
	reflect.TypeOf(new(GuildReadyEvent)):        gateway.IntentGuilds,
	reflect.TypeOf(new(GuildAvailableEvent)):    gateway.IntentGuilds,
	reflect.TypeOf(new(GuildJoinEvent)):         gateway.IntentGuilds,
	reflect.TypeOf(new(GuildUpdateEvent)):       gateway.IntentGuilds,
	reflect.TypeOf(new(GuildDeleteEvent)):       gateway.IntentGuilds,
	reflect.TypeOf(new(GuildUnavailableEvent)):  gateway.IntentGuilds,
	reflect.TypeOf(new(GuildLeaveEvent)):        gateway.IntentGuilds,
	reflect.TypeOf(new(GuildRoleCreateEvent)):   gateway.IntentGuilds,
	reflect.TypeOf(new(GuildRoleUpdateEvent)):   gateway.IntentGuilds,
	reflect.TypeOf(new(GuildRoleDeleteEvent)):   gateway.IntentGuilds,
	reflect.TypeOf(new(ChannelCreateEvent)):     gateway.IntentGuilds,
	reflect.TypeOf(new(ChannelUpdateEvent)):     gateway.IntentGuilds,
	reflect.TypeOf(new(ChannelDeleteEvent)):     gateway.IntentGuilds,
	reflect.TypeOf(new(ChannelPinsUpdateEvent)): gateway.IntentGuilds | gateway.IntentDirectMessages,

	reflect.TypeOf(new(GuildMemberAddEvent)):    gateway.IntentGuildMembers,
	reflect.TypeOf(new(GuildMemberRemoveEvent)): gateway.IntentGuildMembers,
	reflect.TypeOf(new(GuildMemberUpdateEvent)): gateway.IntentGuildMembers,

	reflect.TypeOf(new(GuildBanAddEvent)):    gateway.IntentGuildBans,
	reflect.TypeOf(new(GuildBanRemoveEvent)): gateway.IntentGuildBans,

	reflect.TypeOf(new(GuildEmojisUpdateEvent)): gateway.IntentGuildEmojis,

	reflect.TypeOf(new(GuildIntegrationsUpdateEvent)): gateway.IntentGuildIntegrations,

	reflect.TypeOf(new(WebhooksUpdateEvent)): gateway.IntentGuildWebhooks,

	reflect.TypeOf(new(InviteCreateEvent)): gateway.IntentGuildInvites,
	reflect.TypeOf(new(InviteDeleteEvent)): gateway.IntentGuildInvites,

	reflect.TypeOf(new(VoiceStateUpdateEvent)): gateway.IntentGuildVoiceStates,

	reflect.TypeOf(new(PresenceUpdateEvent)): gateway.IntentGuildPresences,

	reflect.TypeOf(new(MessageCreateEvent)):     gateway.IntentGuildMessages | gateway.IntentDirectMessages,
	reflect.TypeOf(new(MessageUpdateEvent)):     gateway.IntentGuildMessages | gateway.IntentDirectMessages,
	reflect.TypeOf(new(MessageDeleteEvent)):     gateway.IntentGuildMessages | gateway.IntentDirectMessages,
	reflect.TypeOf(new(MessageDeleteBulkEvent)): gateway.IntentGuildMessages,

	reflect.TypeOf(new(MessageReactionAddEvent)): gateway.IntentGuildMessageReactions |
		gateway.IntentDirectMessageReactions,
	reflect.TypeOf(new(MessageReactionRemoveEvent)): gateway.IntentGuildMessageReactions |
		gateway.IntentDirectMessageReactions,
	reflect.TypeOf(new(MessageReactionRemoveAllEvent)): gateway.IntentGuildMessageReactions |
		gateway.IntentDirectMessageReactions,
	reflect.TypeOf(new(MessageReactionRemoveEmojiEvent)): gateway.IntentGuildMessageReactions |
		gateway.IntentDirectMessageReactions,

	reflect.TypeOf(new(TypingStartEvent)): gateway.IntentGuildMessageTyping | gateway.IntentDirectMessageTyping,
}
