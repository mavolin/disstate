package state

import "github.com/diamondburned/arikawa/gateway"

// ================================ Webhooks Update ================================

// https://discord.com/developers/docs/topics/gateway#webhooks-update
type WebhooksUpdateEvent struct {
	*gateway.WebhooksUpdateEvent
	*Base
}

func (e *WebhooksUpdateEvent) getType() eventType { return eventTypeWebhooksUpdate }

type webhooksUpdateEventHandler func(s *State, e *WebhooksUpdateEvent) error

func (h webhooksUpdateEventHandler) handle(s *State, e interface{}) error {
	if e, ok := e.(*WebhooksUpdateEvent); ok {
		return h(s, e)
	}

	return nil
}
