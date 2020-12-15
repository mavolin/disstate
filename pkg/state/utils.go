package state

import (
	"reflect"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

// handleResult handles the passed result of a handler func.
func (h *EventHandler) handleResult(res []reflect.Value) bool {
	if len(res) == 0 {
		return false
	}

	err := res[0].Interface()
	if err == Filtered {
		return true
	} else if err != nil {
		h.ErrorHandler(err.(error))
		return true
	}

	return false
}

// genEvent generates a disstate event from the passed arikawa event.
func (h *EventHandler) genEvent(src interface{}) interface{} {
	base := NewBase()

	switch src := src.(type) {
	// ---------------- Ready Event ----------------
	case *gateway.ReadyEvent:
		return &ReadyEvent{
			ReadyEvent: src,
			Base:       base,
		}

	// ---------------- Channel Events ----------------
	case *gateway.ChannelCreateEvent:
		return &ChannelCreateEvent{
			ChannelCreateEvent: src,
			Base:               base,
		}
	case *gateway.ChannelUpdateEvent:
		c, _ := h.s.Store.Channel(src.ID)

		return &ChannelUpdateEvent{
			ChannelUpdateEvent: src,
			Base:               base,
			Old:                c,
		}
	case *gateway.ChannelDeleteEvent:
		c, _ := h.s.Store.Channel(src.ID)

		return &ChannelDeleteEvent{
			ChannelDeleteEvent: src,
			Base:               base,
			Old:                c,
		}
	case *gateway.ChannelPinsUpdateEvent:
		return &ChannelPinsUpdateEvent{
			ChannelPinsUpdateEvent: src,
			Base:                   base,
		}
	case *gateway.ChannelUnreadUpdateEvent:
		return &ChannelUnreadUpdateEvent{
			ChannelUnreadUpdateEvent: src,
			Base:                     base,
		}

	// ---------------- Guild Events ----------------
	case *gateway.GuildCreateEvent:
		return &GuildCreateEvent{
			GuildCreateEvent: src,
			Base:             base,
		}
	case *gateway.GuildUpdateEvent:
		g, _ := h.s.Store.Guild(src.ID)

		return &GuildUpdateEvent{
			GuildUpdateEvent: src,
			Base:             base,
			Old:              g,
		}
	case *gateway.GuildDeleteEvent:
		g, _ := h.s.Store.Guild(src.ID)

		return &GuildDeleteEvent{
			GuildDeleteEvent: src,
			Base:             base,
			Old:              g,
		}
	case *gateway.GuildBanAddEvent:
		return &GuildBanAddEvent{
			GuildBanAddEvent: src,
			Base:             base,
		}
	case *gateway.GuildBanRemoveEvent:
		return &GuildBanRemoveEvent{
			GuildBanRemoveEvent: src,
			Base:                base,
		}
	case *gateway.GuildEmojisUpdateEvent:
		e, _ := h.s.Store.Emojis(src.GuildID)

		return &GuildEmojisUpdateEvent{
			GuildEmojisUpdateEvent: src,
			Base:                   base,
			Old:                    e,
		}
	case *gateway.GuildIntegrationsUpdateEvent:
		return &GuildIntegrationsUpdateEvent{
			GuildIntegrationsUpdateEvent: src,
			Base:                         base,
		}
	case *gateway.GuildMemberAddEvent:
		return &GuildMemberAddEvent{
			GuildMemberAddEvent: src,
			Base:                base,
		}
	case *gateway.GuildMemberRemoveEvent:
		m, _ := h.s.Store.Member(src.GuildID, src.User.ID)

		return &GuildMemberRemoveEvent{
			GuildMemberRemoveEvent: src,
			Base:                   base,
			Old:                    m,
		}
	case *gateway.GuildMemberUpdateEvent:
		m, _ := h.s.Store.Member(src.GuildID, src.User.ID)

		return &GuildMemberUpdateEvent{
			GuildMemberUpdateEvent: src,
			Base:                   base,
			Old:                    m,
		}
	case *gateway.GuildMembersChunkEvent:
		return &GuildMembersChunkEvent{
			GuildMembersChunkEvent: src,
			Base:                   base,
		}
	case *gateway.GuildRoleCreateEvent:
		return &GuildRoleCreateEvent{
			GuildRoleCreateEvent: src,
			Base:                 base,
		}
	case *gateway.GuildRoleUpdateEvent:
		r, _ := h.s.Store.Role(src.GuildID, src.Role.ID)

		return &GuildRoleUpdateEvent{
			GuildRoleUpdateEvent: src,
			Base:                 base,
			Old:                  r,
		}
	case *gateway.GuildRoleDeleteEvent:
		r, _ := h.s.Store.Role(src.GuildID, src.RoleID)

		return &GuildRoleDeleteEvent{
			GuildRoleDeleteEvent: src,
			Base:                 base,
			Old:                  r,
		}

	// ---------------- Invite Events ----------------
	case *gateway.InviteCreateEvent:
		return &InviteCreateEvent{
			InviteCreateEvent: src,
			Base:              base,
		}
	case *gateway.InviteDeleteEvent:
		return &InviteDeleteEvent{
			InviteDeleteEvent: src,
			Base:              base,
		}

	// ---------------- Message Events ----------------
	case *gateway.MessageCreateEvent:
		return &MessageCreateEvent{
			MessageCreateEvent: src,
			Base:               base,
		}
	case *gateway.MessageUpdateEvent:
		m, _ := h.s.Store.Message(src.ChannelID, src.ID)

		return &MessageUpdateEvent{
			MessageUpdateEvent: src,
			Base:               base,
			Old:                m,
		}
	case *gateway.MessageDeleteEvent:
		m, _ := h.s.Store.Message(src.ChannelID, src.ID)

		return &MessageDeleteEvent{
			MessageDeleteEvent: src,
			Base:               base,
			Old:                m,
		}
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
		}
	case *gateway.MessageReactionAddEvent:
		return &MessageReactionAddEvent{
			MessageReactionAddEvent: src,
			Base:                    base,
		}
	case *gateway.MessageReactionRemoveEvent:
		return &MessageReactionRemoveEvent{
			MessageReactionRemoveEvent: src,
			Base:                       base,
		}
	case *gateway.MessageReactionRemoveAllEvent:
		return &MessageReactionRemoveAllEvent{
			MessageReactionRemoveAllEvent: src,
			Base:                          base,
		}
	case *gateway.MessageReactionRemoveEmoji:
		return &MessageReactionRemoveEmojiEvent{
			MessageReactionRemoveEmoji: src,
			Base:                       base,
		}
	case *gateway.MessageAckEvent:
		return &MessageAckEvent{
			MessageAckEvent: src,
			Base:            base,
		}

	// ---------------- Presence Events ----------------
	case *gateway.PresenceUpdateEvent:
		p, _ := h.s.Store.Presence(src.GuildID, src.User.ID)

		return &PresenceUpdateEvent{
			PresenceUpdateEvent: src,
			Base:                base,
			Old:                 p,
		}
	case *gateway.PresencesReplaceEvent:
		return &PresencesReplaceEvent{
			PresencesReplaceEvent: src,
			Base:                  base,
		}
	case *gateway.SessionsReplaceEvent:
		return &SessionsReplaceEvent{
			SessionsReplaceEvent: src,
			Base:                 base,
		}
	case *gateway.TypingStartEvent:
		return &TypingStartEvent{
			TypingStartEvent: src,
			Base:             base,
		}
	case *gateway.UserUpdateEvent:
		return &UserUpdateEvent{
			UserUpdateEvent: src,
			Base:            base,
		}

	// ---------------- Relationship Events ----------------
	case *gateway.RelationshipAddEvent:
		return &RelationshipAddEvent{
			RelationshipAddEvent: src,
			Base:                 base,
		}
	case *gateway.RelationshipRemoveEvent:
		return &RelationshipRemoveEvent{
			RelationshipRemoveEvent: src,
			Base:                    base,
		}

	// ---------------- User Settings Events ----------------
	case *gateway.UserGuildSettingsUpdateEvent:
		return &UserGuildSettingsUpdateEvent{
			UserGuildSettingsUpdateEvent: src,
			Base:                         base,
		}
	case *gateway.UserSettingsUpdateEvent:
		return &UserSettingsUpdateEvent{
			UserSettingsUpdateEvent: src,
			Base:                    base,
		}
	case *gateway.UserNoteUpdateEvent:
		return &UserNoteUpdateEvent{
			UserNoteUpdateEvent: src,
			Base:                base,
		}

	// ---------------- Voice Events ----------------
	case *gateway.VoiceStateUpdateEvent:
		return &VoiceStateUpdateEvent{
			VoiceStateUpdateEvent: src,
			Base:                  base,
		}
	case *gateway.VoiceServerUpdateEvent:
		return &VoiceServerUpdateEvent{
			VoiceServerUpdateEvent: src,
			Base:                   base,
		}

	// ---------------- Webhook Events ----------------
	case *gateway.WebhooksUpdateEvent:
		return &WebhooksUpdateEvent{
			WebhooksUpdateEvent: src,
			Base:                base,
		}
	}

	return nil
}

// copyEvent copies the event stored in the passed reflect.Value with the
// passed reflect.Type.
// v must not be a pointer however, t is expected to be the pointerized type
// of v.
func copyEvent(v reflect.Value, t reflect.Type) reflect.Value {
	cp := reflect.New(t.Elem())
	cp = cp.Elem()

	for i := 0; i < v.NumField(); i++ {
		cp.Field(i).Set(v.Field(i))
	}

	b := v.FieldByName("Base").Interface().(*Base)
	bcp := b.copy()

	bcpValue := reflect.ValueOf(bcp)

	cp.FieldByName("Base").Set(bcpValue)

	return cp.Addr()
}
