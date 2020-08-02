package state

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/state"
	"github.com/pkg/errors"
)

func (s *State) updateStore(e interface{}) {
	switch e := e.(type) {
	case *gateway.ReadyEvent:
		s.Ready = e

		// Handle presences
		for _, p := range e.Presences {
			if err := s.Store.PresenceSet(0, p); err != nil {
				s.stateErr(err, "failed to set global presence")
			}
		}

		// Handle guilds
		for i := range e.Guilds {
			s.batchLog(storeGuildCreate(s.Store, &e.Guilds[i])...)
		}

		// Handle private channels
		for _, ch := range e.PrivateChannels {
			if err := s.Store.ChannelSet(ch); err != nil {
				s.stateErr(err, "failed to set channel in state")
			}
		}

		// Handle user
		if err := s.Store.MyselfSet(e.User); err != nil {
			s.stateErr(err, "failed to set self in state")
		}

	case *gateway.GuildUpdateEvent:
		if err := s.Store.GuildSet(e.Guild); err != nil {
			s.stateErr(err, "failed to update guild in state")
		}

	case *gateway.GuildDeleteEvent:
		if err := s.Store.GuildRemove(e.ID); err != nil && !e.Unavailable {
			s.stateErr(err, "failed to delete guild in state")
		}

	case *gateway.GuildMemberAddEvent:
		if err := s.Store.MemberSet(e.GuildID, e.Member); err != nil {
			s.stateErr(err, "failed to add a member in state")
		}

	case *gateway.GuildMemberUpdateEvent:
		m, err := s.Store.Member(e.GuildID, e.User.ID)
		if err != nil {
			// We can't do much here.
			m = &discord.Member{}
		}

		// Update available fields from e into m
		e.Update(m)

		if err := s.Store.MemberSet(e.GuildID, *m); err != nil {
			s.stateErr(err, "failed to update a member in state")
		}

	case *gateway.GuildMemberRemoveEvent:
		if err := s.Store.MemberRemove(e.GuildID, e.User.ID); err != nil {
			s.stateErr(err, "failed to remove a member in state")
		}

	case *gateway.GuildMembersChunkEvent:
		for _, m := range e.Members {
			if err := s.Store.MemberSet(e.GuildID, m); err != nil {
				s.stateErr(err, "failed to add a member from chunk in state")
			}
		}

		for _, p := range e.Presences {
			if err := s.Store.PresenceSet(e.GuildID, p); err != nil {
				s.stateErr(err, "failed to add a presence from chunk in state")
			}
		}

	case *gateway.GuildRoleCreateEvent:
		if err := s.Store.RoleSet(e.GuildID, e.Role); err != nil {
			s.stateErr(err, "failed to add a role in state")
		}

	case *gateway.GuildRoleUpdateEvent:
		if err := s.Store.RoleSet(e.GuildID, e.Role); err != nil {
			s.stateErr(err, "failed to update a role in state")
		}

	case *gateway.GuildRoleDeleteEvent:
		if err := s.Store.RoleRemove(e.GuildID, e.RoleID); err != nil {
			s.stateErr(err, "failed to remove a role in state")
		}

	case *gateway.GuildEmojisUpdateEvent:
		if err := s.Store.EmojiSet(e.GuildID, e.Emojis); err != nil {
			s.stateErr(err, "failed to update emojis in state")
		}

	case *gateway.ChannelCreateEvent:
		if err := s.Store.ChannelSet(e.Channel); err != nil {
			s.stateErr(err, "failed to create a channel in state")
		}

	case *gateway.ChannelUpdateEvent:
		if err := s.Store.ChannelSet(e.Channel); err != nil {
			s.stateErr(err, "failed to update a channel in state")
		}

	case *gateway.ChannelDeleteEvent:
		if err := s.Store.ChannelRemove(e.Channel); err != nil {
			s.stateErr(err, "failed to remove a channel in state")
		}

	case *gateway.ChannelPinsUpdateEvent:
		// not tracked.

	case *gateway.MessageCreateEvent:
		if err := s.Store.MessageSet(e.Message); err != nil {
			s.stateErr(err, "failed to add a message in state")
		}

	case *gateway.MessageUpdateEvent:
		if err := s.Store.MessageSet(e.Message); err != nil {
			s.stateErr(err, "failed to update a message in state")
		}

	case *gateway.MessageDeleteEvent:
		if err := s.Store.MessageRemove(e.ChannelID, e.ID); err != nil {
			s.stateErr(err, "failed to delete a message in state")
		}

	case *gateway.MessageDeleteBulkEvent:
		for _, id := range e.IDs {
			if err := s.Store.MessageRemove(e.ChannelID, id); err != nil {
				s.stateErr(err, "failed to delete bulk messages in state")
			}
		}

	case *gateway.MessageReactionAddEvent:
		s.editMessage(e.ChannelID, e.MessageID, func(m *discord.Message) bool {
			if i := findReaction(m.Reactions, e.Emoji); i > -1 {
				m.Reactions[i].Count++
			} else {
				var me bool
				if u, _ := s.Store.Me(); u != nil {
					me = e.UserID == u.ID
				}
				m.Reactions = append(m.Reactions, discord.Reaction{
					Count: 1,
					Me:    me,
					Emoji: e.Emoji,
				})
			}
			return true
		})

	case *gateway.MessageReactionRemoveEvent:
		s.editMessage(e.ChannelID, e.MessageID, func(m *discord.Message) bool {
			var i = findReaction(m.Reactions, e.Emoji)
			if i < 0 {
				return false
			}

			r := &m.Reactions[i]
			r.Count--

			switch {
			case r.Count < 1: // If the count is 0:
				// Remove the reaction.
				m.Reactions = append(m.Reactions[:i], m.Reactions[i+1:]...)

			case r.Me: // If reaction removal is the user's
				u, err := s.Store.Me()
				if err == nil && e.UserID == u.ID {
					r.Me = false
				}
			}

			return true
		})

	case *gateway.MessageReactionRemoveAllEvent:
		s.editMessage(e.ChannelID, e.MessageID, func(m *discord.Message) bool {
			m.Reactions = nil
			return true
		})

	case *gateway.MessageReactionRemoveEmoji:
		s.editMessage(e.ChannelID, e.MessageID, func(m *discord.Message) bool {
			var i = findReaction(m.Reactions, e.Emoji)
			if i < 0 {
				return false
			}
			m.Reactions = append(m.Reactions[:i], m.Reactions[i+1:]...)
			return true
		})

	case *gateway.PresenceUpdateEvent:
		if err := s.Store.PresenceSet(e.GuildID, e.Presence); err != nil {
			s.stateErr(err, "failed to update presence in state")
		}

	case *gateway.PresencesReplaceEvent:
		for _, p := range *e {
			if err := s.Store.PresenceSet(p.GuildID, p); err != nil {
				s.stateErr(err, "failed to update presence in state")
			}
		}

	case *gateway.SessionsReplaceEvent:

	case *gateway.UserGuildSettingsUpdateEvent:
		for i, ugs := range s.Ready.UserGuildSettings {
			if ugs.GuildID == e.GuildID {
				s.Ready.UserGuildSettings[i] = e.UserGuildSettings
			}
		}

	case *gateway.UserSettingsUpdateEvent:
		s.Ready.Settings = &e.UserSettings

	case *gateway.UserNoteUpdateEvent:
		s.Ready.Notes[e.ID] = e.Note

	case *gateway.UserUpdateEvent:
		if err := s.Store.MyselfSet(e.User); err != nil {
			s.stateErr(err, "failed to update myself from USER_UPDATE")
		}

	case *gateway.VoiceStateUpdateEvent:
		vs := &e.VoiceState
		if vs.ChannelID == 0 {
			if err := s.Store.VoiceStateRemove(vs.GuildID, vs.UserID); err != nil {
				s.stateErr(err, "failed to remove voice state from state")
			}
		} else {
			if err := s.Store.VoiceStateSet(vs.GuildID, *vs); err != nil {
				s.stateErr(err, "failed to update voice state in state")
			}
		}
	}
}

func (s *State) stateErr(err error, wrap string) {
	s.StateLog(errors.Wrap(err, wrap))
}
func (s *State) batchLog(errors ...error) {
	for _, err := range errors {
		s.StateLog(err)
	}
}

func (s *State) editMessage(ch discord.ChannelID, msg discord.MessageID, fn func(m *discord.Message) bool) {
	m, err := s.Store.Message(ch, msg)
	if err != nil {
		return
	}
	if !fn(m) {
		return
	}
	if err := s.Store.MessageSet(*m); err != nil {
		s.stateErr(err, "failed to save message in reaction add")
	}
}

func findReaction(rs []discord.Reaction, emoji discord.Emoji) int {
	for i := range rs {
		if rs[i].Emoji.ID == emoji.ID && rs[i].Emoji.Name == emoji.Name {
			return i
		}
	}
	return -1
}

func storeGuildCreate(store state.Store, guild *gateway.GuildCreateEvent) []error {
	if guild.Unavailable {
		return nil
	}

	stack, errs := newErrorStack()

	if err := store.GuildSet(guild.Guild); err != nil {
		errs(err, "failed to set guild in Ready")
	}

	// Handle guild emojis
	if guild.Emojis != nil {
		if err := store.EmojiSet(guild.ID, guild.Emojis); err != nil {
			errs(err, "failed to set guild emojis")
		}
	}

	// Handle guild member
	for _, m := range guild.Members {
		if err := store.MemberSet(guild.ID, m); err != nil {
			errs(err, "failed to set guild member in Ready")
		}
	}

	// Handle guild channels
	for _, ch := range guild.Channels {
		// I HATE Discord.
		ch.GuildID = guild.ID

		if err := store.ChannelSet(ch); err != nil {
			errs(err, "failed to set guild channel in Ready")
		}
	}

	// Handle guild presences
	for _, p := range guild.Presences {
		if err := store.PresenceSet(guild.ID, p); err != nil {
			errs(err, "failed to set guild presence in Ready")
		}
	}

	// Handle guild voice states
	for _, v := range guild.VoiceStates {
		if err := store.VoiceStateSet(guild.ID, v); err != nil {
			errs(err, "failed to set guild voice state in Ready")
		}
	}

	return *stack
}

func newErrorStack() (*[]error, func(error, string)) {
	var errs = new([]error)
	return errs, func(err error, wrap string) {
		*errs = append(*errs, errors.Wrap(err, wrap))
	}
}
