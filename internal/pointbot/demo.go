package pointbot

import (
	"github.com/devnull-twitch/go-tmi"
)

func DemoCmd() tmi.Command {
	return tmi.Command{
		Name:                     "demo",
		Description:              "Print demo message",
		RequiresBroadcasterOrMod: true,
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			return &tmi.OutgoingMessage{
				Message: "Hello there",
			}
		},
	}
}
