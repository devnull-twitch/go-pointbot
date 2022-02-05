package pointbot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

type pointModule struct {
	storageReqChannel chan<- StorageRequest
}

func (p *pointModule) ExternalTrigger(client *tmi.Client) <-chan *tmi.ModuleArgs {
	return nil
}

func (p *pointModule) MessageTrigger(client *tmi.Client, incoming *tmi.IncomingMessage) *tmi.ModuleArgs {
	return nil
}

func (p *pointModule) Handler(client *tmi.Client, args tmi.ModuleArgs) *tmi.OutgoingMessage {
	return nil
}

func PointModule(storageReqChannel chan<- StorageRequest) tmi.Module {
	return &pointModule{
		storageReqChannel: storageReqChannel,
	}
}

func PointModuleCommand() tmi.ModuleCommand {
	return tmi.ModuleCommand{
		ModuleCommandHandler: func(client *tmi.Client, m tmi.Module, args tmi.CommandArgs) *tmi.OutgoingMessage {
			pm := m.(*pointModule)
			if args.Parameters["sub"] != "" {
				if args.Mod || args.Broadcaster {
					switch args.Parameters["sub"] {
					case "add":
						return pm.addPoints(args.Channel, args.Parameters["user"], args.Parameters["points"])
					case "get":
						return pm.getPoints(args.Channel, args.Parameters["user"])
					case "del":
						return pm.delPoints(args.Channel, args.Parameters["user"])
					case "sub":
						return pm.subPoints(args.Channel, args.Parameters["user"], args.Parameters["points"])
					case "reward":
						// user param is points here because username is "@username" at the start
						return pm.addPoints(args.Channel, args.ReplyUsername, args.Parameters["user"])
					}
				}

				switch args.Parameters["sub"] {
				case "top":
					return pm.topPoints(args.Channel)
				}

				return &tmi.OutgoingMessage{Message: "Unknown sub command"}
			}

			return pm.getPoints(args.Channel, args.Username)
		},
		Command: tmi.Command{
			Name:        "points",
			Description: "Interact with points",
			Params: []tmi.Parameter{
				{Name: "sub"},
				{Name: "user"},
				{Name: "points"},
			},
		},
	}
}

func (pm *pointModule) getPoints(channel, username string) *tmi.OutgoingMessage {
	replychan := make(chan StorageResponse)

	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionGetPoints,
		ChannelName: channel,
		Username:    username,
		ReplyChan:   replychan,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
		return nil
	}

	var response StorageResponse
	select {
	case response = <-replychan:
	case <-time.After(time.Second):
		logrus.Error("storage response timed out")
		return nil
	}

	return &tmi.OutgoingMessage{
		Message: fmt.Sprintf("%s has %d points", username, response.Points),
	}
}

func (pm *pointModule) addPoints(channel, username, strPoints string) *tmi.OutgoingMessage {
	points, err := strconv.Atoi(strPoints)
	if err != nil {
		return &tmi.OutgoingMessage{
			Message: "Unable to read points NotLikeThis",
		}
	}
	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionAddPoints,
		ChannelName: channel,
		Username:    username,
		Points:      points,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
	}

	return &tmi.OutgoingMessage{
		Message: "Ok. Added points!",
	}
}

func (pm *pointModule) subPoints(channel, username, strPoints string) *tmi.OutgoingMessage {
	points, err := strconv.Atoi(strPoints)
	if err != nil {
		return &tmi.OutgoingMessage{
			Message: "Unable to read points NotLikeThis",
		}
	}
	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionSubPoints,
		ChannelName: channel,
		Username:    username,
		Points:      points,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
	}

	return &tmi.OutgoingMessage{
		Message: "Ok. Subtracted some points!",
	}
}

func (pm *pointModule) delPoints(channel, username string) *tmi.OutgoingMessage {
	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionDelPoints,
		ChannelName: channel,
		Username:    username,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
	}

	return &tmi.OutgoingMessage{
		Message: "Ok. Deleted all points!",
	}
}

func (pm *pointModule) topPoints(channel string) *tmi.OutgoingMessage {
	replychan := make(chan StorageResponse)

	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionTop,
		ChannelName: channel,
		ReplyChan:   replychan,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
		return nil
	}

	var response StorageResponse
	select {
	case response = <-replychan:
	case <-time.After(time.Second):
		logrus.Error("storage response timed out")
		return nil
	}

	return &tmi.OutgoingMessage{
		Message: fmt.Sprintf("Top scorer is %s with %d points", response.Username, response.Points),
	}
}
