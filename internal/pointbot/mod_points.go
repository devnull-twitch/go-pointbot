package pointbot

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

type pointModule struct {
	lastLeaderboard   time.Time
	storageReqChannel chan<- StorageRequest
}

func (pm *pointModule) ExternalTrigger(client *tmi.Client) <-chan *tmi.ModuleArgs {
	return nil
}

func (pm *pointModule) MessageTrigger(client *tmi.Client, incoming *tmi.IncomingMessage) *tmi.ModuleArgs {
	if incoming.Message[0:1] == os.Getenv("COMMAND_MARK") {
		return nil
	}

	pm.storageReqChannel <- StorageRequest{
		Action:      ActionChatBasePoint,
		ChannelName: incoming.Channel,
		Username:    incoming.Username,
	}
	return nil
}

func (pm *pointModule) Handler(client *tmi.Client, args tmi.ModuleArgs) *tmi.OutgoingMessage {
	return nil
}

func PointModule(storageReqChannel chan<- StorageRequest) tmi.Module {
	return &pointModule{
		storageReqChannel: storageReqChannel,
		lastLeaderboard:   time.Now().Add(-(time.Minute * 5)),
	}
}

func LeaderboardCommand(m tmi.Module) tmi.Command {
	pm := m.(*pointModule)
	defaultLength := "3"
	return tmi.Command{
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			if pm.lastLeaderboard.After(time.Now().Add(-(time.Minute))) {
				logrus.Info("no spwam leaderboard")
				return nil
			}
			pm.lastLeaderboard = time.Now()

			replyChan := make(chan StorageResponse)
			select {
			case pm.storageReqChannel <- StorageRequest{
				Action:      ActionList,
				ChannelName: args.Channel,
				ReplyChan:   replyChan,
			}:
			case <-time.After(time.Second):
			}

			max := 3
			userm, err := strconv.Atoi(args.Parameters["length"])
			if err == nil && userm < 5 {
				max = userm
			}

			var i int
			goon := true
			for goon {
				select {
				case r, ok := <-replyChan:
					if !ok {
						goon = false
						break
					}

					if i+1 <= max {
						client.Send(&tmi.OutgoingMessage{
							Channel: args.Channel,
							Message: fmt.Sprintf("%d) %d points %s", i+1, r.Points, r.Username),
						})
					}
					i++

				case <-time.After(time.Second):
					goon = false
				}
			}

			return nil
		},
		Name:        "leaderboard",
		Description: "Show top 3",
		Params: []tmi.Parameter{
			{Name: "length", Default: &defaultLength},
		},
	}
}

func PointModuleCommand(m tmi.Module, ffc FeatureFlagChecker) tmi.Command {
	pm := m.(*pointModule)
	return tmi.Command{
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
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
				case "gift":
					return pm.giftPoints(args.Channel, args.Username, args.Parameters["user"], args.Parameters["points"])
				}

				return &tmi.OutgoingMessage{Message: "Unknown sub command"}
			}

			return pm.getPoints(args.Channel, args.Username)
		},
		Name:            "points",
		Description:     "Interact with points",
		AcceptanceCheck: ffc.FeatureFlagAccptanceCheck,
		Params: []tmi.Parameter{
			{Name: "sub"},
			{Name: "user"},
			{Name: "points"},
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

func (pm *pointModule) giftPoints(channel, providerUsername, receiverUsername, strPoints string) *tmi.OutgoingMessage {
	points, err := strconv.Atoi(strPoints)
	if err != nil {
		return &tmi.OutgoingMessage{
			Message: "Unable to read points NotLikeThis",
		}
	}

	wasNegative := false
	if points <= 0 {
		wasNegative = true
		points = int(math.Abs(float64(points)))
	}

	replychan := make(chan StorageResponse)
	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionGetPoints,
		ChannelName: channel,
		Username:    providerUsername,
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

	if response.Points < int64(points) {
		if wasNegative {
			points = int(response.Points)
		} else {
			return &tmi.OutgoingMessage{
				Message: "You do not have enough points NotLikeThis",
			}
		}
	}

	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionSubPoints,
		ChannelName: channel,
		Username:    providerUsername,
		Points:      points,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
		return nil
	}

	if wasNegative {
		return &tmi.OutgoingMessage{
			Message: fmt.Sprintf("Trashed %d points from %s", points, providerUsername),
		}
	}

	select {
	case pm.storageReqChannel <- StorageRequest{
		Action:      ActionAddPoints,
		ChannelName: channel,
		Username:    receiverUsername,
		Points:      points,
	}:
	case <-time.After(time.Second):
		logrus.Error("storage request timed out")
		return nil
	}

	return &tmi.OutgoingMessage{
		Message: fmt.Sprintf("Transfered %d points from %s to %s", points, providerUsername, receiverUsername),
	}
}
