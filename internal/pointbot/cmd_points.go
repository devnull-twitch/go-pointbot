package pointbot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

func ShowPoints(storageReqChannel chan<- StorageRequest) tmi.Command {
	return tmi.Command{
		Name:        "points",
		Description: "Interact with points",
		Params: []tmi.Parameter{
			{Name: "sub"},
			{Name: "user"},
			{Name: "points"},
		},
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			if args.Parameters["sub"] != "" {
				if args.UserIsBroadcasterOrMod {
					switch args.Parameters["sub"] {
					case "add":
						return addPoints(storageReqChannel, args.Channel, args.Parameters["user"], args.Parameters["points"])
					case "get":
						return getPoints(storageReqChannel, args.Channel, args.Parameters["user"])
					case "del":
						return delPoints(storageReqChannel, args.Channel, args.Parameters["user"])
					case "sub":
						return subPoints(storageReqChannel, args.Channel, args.Parameters["user"], args.Parameters["points"])
					}
				}

				switch args.Parameters["sub"] {
				case "top":
					return topPoints(storageReqChannel, args.Channel)
				}

				return nil
			}

			return getPoints(storageReqChannel, args.Channel, args.Username)
		},
	}
}

func getPoints(stoargeReqChannel chan<- StorageRequest, channel, username string) *tmi.OutgoingMessage {
	replychan := make(chan StorageResponse)

	select {
	case stoargeReqChannel <- StorageRequest{
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

func addPoints(stoargeReqChannel chan<- StorageRequest, channel, username, strPoints string) *tmi.OutgoingMessage {
	points, err := strconv.Atoi(strPoints)
	if err != nil {
		return &tmi.OutgoingMessage{
			Message: "Unable to read points NotLikeThis",
		}
	}
	select {
	case stoargeReqChannel <- StorageRequest{
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

func subPoints(stoargeReqChannel chan<- StorageRequest, channel, username, strPoints string) *tmi.OutgoingMessage {
	points, err := strconv.Atoi(strPoints)
	if err != nil {
		return &tmi.OutgoingMessage{
			Message: "Unable to read points NotLikeThis",
		}
	}
	select {
	case stoargeReqChannel <- StorageRequest{
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

func delPoints(stoargeReqChannel chan<- StorageRequest, channel, username string) *tmi.OutgoingMessage {
	select {
	case stoargeReqChannel <- StorageRequest{
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

func topPoints(stoargeReqChannel chan<- StorageRequest, channel string) *tmi.OutgoingMessage {
	replychan := make(chan StorageResponse)

	select {
	case stoargeReqChannel <- StorageRequest{
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
