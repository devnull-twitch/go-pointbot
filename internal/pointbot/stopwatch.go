package pointbot

import (
	"fmt"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

type stopwatchUserInfo struct {
	Channel   string
	LastStart time.Time
	Running   bool
}

var playerCache = map[string][]stopwatchUserInfo{}

func StopwatchCommand() tmi.Command {
	return tmi.Command{
		Name: "stopwatch",
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			fmt.Println("stopwatch command handler")
			infos, ok := playerCache[args.Username]
			if !ok {
				return startStopwatch(args)
			}

			for _, info := range infos {
				if info.Channel == args.Channel && info.Running {
					return stopStopwatch(args)
				} else {
					if time.Since(info.LastStart) < time.Hour && !args.Broadcaster && !args.Mod {
						logrus.WithFields(logrus.Fields{
							"user":    args.Username,
							"channel": args.Channel,
						}).Warn("user liked stopwatch a bit too much")
						return nil
					}
					return startStopwatch(args)
				}
			}

			return startStopwatch(args)
		},
	}
}

func startStopwatch(args tmi.CommandArgs) *tmi.OutgoingMessage {
	// cleanup old entries
	for _, info := range playerCache[args.Username] {
		filtered := make([]stopwatchUserInfo, 0, len(playerCache[args.Username]))
		if info.Channel != args.Channel {
			filtered = append(filtered, info)
		}
		playerCache[args.Username] = filtered
	}

	logrus.WithFields(logrus.Fields{
		"channel": args.Channel,
		"user":    args.Username,
	}).Info("started stopwatch")

	playerCache[args.Username] = append(playerCache[args.Username], stopwatchUserInfo{
		Channel:   args.Channel,
		Running:   true,
		LastStart: time.Now(),
	})
	return &tmi.OutgoingMessage{
		Channel: args.Channel,
		Message: "Stopwatch started",
	}
}

func stopStopwatch(args tmi.CommandArgs) *tmi.OutgoingMessage {
	for _, info := range playerCache[args.Username] {
		if info.Channel == args.Channel {
			if !info.Running {
				return nil
			}

			dur := time.Since(info.LastStart)
			return &tmi.OutgoingMessage{
				Channel: args.Channel,
				Message: fmt.Sprintf("Stopped after %s", dur),
			}
		}
	}

	return nil
}
