// Package event provides wrappers around arikawa's events as well as a custom
// event handler.
//
// Event Handler
//
// disstate's event handler expands the functionality of arikawa's built-in
// handler system.
// It adds support for middlewares both on a global level and a per-handler
// level.
//
// Furthermore, *state.State is the first parameter of all events, to allow
// access to the state without wrapping functions.
// Optionally, handler functions may also return an error that is handled
// by the event handler's ErrorHandler function.
// Similarly, the event handler will also recover from panics and handle them
// using the event handler's PanicHandler function.
//
// Event Types
//
// Wrapped events have two key differences in comparison to their arikawa
// counterparts.
//
// Firstly, they embed a *Base that contains a key-value-store similar to that
// of context.Context.
// This allows to share state between middlewares and handlers.
//
// Secondly, some events have an Old field, containing the state of event's
// entity prior to when the event was received.
// The Old field will obviously only be filled, if the enitity was previously
// cached.
// This functionality replaces arikawa's PreHandler system and allows handlers
// to work with both the previous and the current entity, something impossible
// with arikawa's system.
package event

import (
	"reflect"

	"github.com/diamondburned/arikawa/v3/gateway"
)

//go:generate go run ../../tools/codegen/event/event.go

var (
	interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	baseType      = reflect.TypeOf(new(Base))

	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

var eventIntents map[reflect.Type]gateway.Intents

func init() {
	eventIntents = make(map[reflect.Type]gateway.Intents, len(gateway.EventIntents))

	for event, intent := range gateway.EventIntents {
		constructor, ok := gateway.EventCreator[event]
		if !ok {
			continue
		}

		eventIntents[reflect.TypeOf(constructor())] = intent
	}
}
