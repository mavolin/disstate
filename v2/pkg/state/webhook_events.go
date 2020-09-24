package state

import "github.com/diamondburned/arikawa/gateway"

// https://discord.com/developers/docs/topics/gateway#webhooks-update
type WebhooksUpdateEvent struct {
	*gateway.WebhooksUpdateEvent
	*Base
}
