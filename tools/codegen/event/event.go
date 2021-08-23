// Package event generates the event wrapper structs and the event generator
// function.
package main

import (
	"embed"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/diamondburned/arikawa/v3/gateway"
)

//go:embed *.gotmpl
var templateFS embed.FS

type (
	// Event is the data known about a single event.
	Event struct {
		// Name is the go type name of the event.
		Name string
		// GatewayName is the name of the type as found in the gateway package.
		GatewayName string
		// Old contains information about the Old field, or is nil if the event
		// has no such field.
		Old *Old

		// Intents are the gateway.Intents required to receive this event.
		Intents gateway.Intents
	}

	// Old contains information about how to fill the old field of an event.
	Old struct {
		// Type is the type of the old field.
		// It's package should be discord.
		Type string
		// CabinetFuncName is the name of the cabinet function to retrieve the
		// old value from.
		CabinetFuncName string
		// CabinetFuncParams are the parameters given to the cabinet function.
		// They should correspond to the names of the fields of the event.
		CabinetFuncParams []string
	}
)

var eventsWithOld = map[string]*Old{
	"CHANNEL_UPDATE": {
		Type:              "*discord.Channel",
		CabinetFuncName:   "Channel",
		CabinetFuncParams: []string{"ID"},
	},
	"CHANNEL_DELETE": {
		Type:              "*discord.Channel",
		CabinetFuncName:   "Channel",
		CabinetFuncParams: []string{"ID"},
	},
	"GUILD_UPDATE": {
		Type:              "*discord.Guild",
		CabinetFuncName:   "Guild",
		CabinetFuncParams: []string{"ID"},
	},
	"GUILD_DELETE": {
		Type:              "*discord.Guild",
		CabinetFuncName:   "Guild",
		CabinetFuncParams: []string{"ID"},
	},
	"GUILD_MEMBER_UPDATE": {
		Type:              "*discord.Member",
		CabinetFuncName:   "Member",
		CabinetFuncParams: []string{"GuildID", "User.ID"},
	},
	"GUILD_MEMBER_REMOVE": {
		Type:              "*discord.Member",
		CabinetFuncName:   "Member",
		CabinetFuncParams: []string{"GuildID", "User.ID"},
	},
	"GUILD_ROLE_UPDATE": {
		Type:              "*discord.Role",
		CabinetFuncName:   "Role",
		CabinetFuncParams: []string{"GuildID", "Role.ID"},
	},
	"GUILD_ROLE_DELETE": {
		Type:              "*discord.Role",
		CabinetFuncName:   "Role",
		CabinetFuncParams: []string{"GuildID", "RoleID"},
	},
	"MESSAGE_UPDATE": {
		Type:              "*discord.Message",
		CabinetFuncName:   "Message",
		CabinetFuncParams: []string{"ChannelID", "ID"},
	},
	"MESSAGE_DELETE": {
		Type:              "*discord.Message",
		CabinetFuncName:   "Message",
		CabinetFuncParams: []string{"ChannelID", "ID"},
	},
	"PRESENCE_UPDATE": {
		Type:              "*discord.Presence",
		CabinetFuncName:   "Presence",
		CabinetFuncParams: []string{"GuildID", "User.ID"},
	},
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	events := make([]Event, 0, len(gateway.EventCreator))

	for discordName, constructor := range gateway.EventCreator {
		event := Event{
			// this is ok to do, since the map will return nil, if there is no
			// entry for the given name
			Old:     eventsWithOld[discordName],
			Intents: gateway.EventIntents[discordName],
		}

		event.GatewayName = reflect.TypeOf(constructor()).Elem().Name()

		event.Name = event.GatewayName
		event.Name = strings.TrimSuffix(event.GatewayName, "Event")

		events = append(events, event)
	}

	log.Printf("extracted %d events\n", len(events))

	if err := generateEvents(events); err != nil {
		return err
	}

	if err := generateGenerator(events); err != nil {
		return err
	}

	log.Println("done")

	return nil
}

func generateEvents(events []Event) error {
	log.Println("generating events.go")

	f, err := os.Create("events.go")
	if err != nil {
		return err
	}

	t, err := template.ParseFS(templateFS, "events.gotmpl")
	if err != nil {
		return err
	}

	return t.Execute(f, events)
}

func generateGenerator(events []Event) error {
	log.Println("generating generator.go")

	f, err := os.Create("generator.go")
	if err != nil {
		return err
	}

	t, err := template.ParseFS(templateFS, "generator.gotmpl")
	if err != nil {
		return err
	}

	return t.Execute(f, events)
}
