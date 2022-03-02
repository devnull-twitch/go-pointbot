package pointbot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

type (
	wager struct {
		value  int
		option string
	}

	gamblingBet struct {
		channel      string
		isOpen       bool
		isChatterBet bool
		options      []string
		wagers       map[string]wager
		method       string
	}

	gamblingMod struct {
		closingChan       chan *gamblingBet
		runningBets       map[string]*gamblingBet
		storageReqChannel chan<- StorageRequest
		client            *tmi.Client
	}
)

func GamblingModule(client *tmi.Client, storageReqChannel chan<- StorageRequest) tmi.Module {
	mod := &gamblingMod{
		storageReqChannel: storageReqChannel,
		runningBets:       make(map[string]*gamblingBet),
		closingChan:       make(chan *gamblingBet),
		client:            client,
	}
	go func() {
		for {
			bet := <-mod.closingChan
			bet.isOpen = false
			client.Send(&tmi.OutgoingMessage{
				Channel: bet.channel,
				Message: "rien ne va plus",
			})
		}
	}()
	return mod
}

func (pm *gamblingMod) ExternalTrigger(client *tmi.Client) <-chan *tmi.ModuleArgs {
	return nil
}

func (pm *gamblingMod) Handler(client *tmi.Client, args tmi.ModuleArgs) *tmi.OutgoingMessage {
	return nil
}

func (pm *gamblingMod) MessageTrigger(client *tmi.Client, incoming *tmi.IncomingMessage) *tmi.ModuleArgs {
	return nil
}

func hasOption(options []string, input string) bool {
	for _, checkOption := range options {
		if checkOption == input {
			return true
		}
	}

	return false
}

func GamblingUserCommand(m tmi.Module) tmi.Command {
	gm := m.(*gamblingMod)
	return tmi.Command{
		Name:        "bet",
		Description: "Place a bet",
		Params: []tmi.Parameter{
			{Name: "amount", Required: true},
			{Name: "option"},
		},
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			currentBet := gm.runningBets[args.Channel]
			if currentBet == nil {
				return &tmi.OutgoingMessage{
					Channel: args.Channel,
					Message: "No active bet BibleThump",
				}
			}

			if !currentBet.isOpen {
				return &tmi.OutgoingMessage{
					Channel:     args.Channel,
					ParentID:    args.MsgID,
					Message:     "Bet is closed",
					SendAsReply: true,
				}
			}

			if _, hasPlacedWager := currentBet.wagers[args.Username]; hasPlacedWager {
				return &tmi.OutgoingMessage{
					Channel:     args.Channel,
					ParentID:    args.MsgID,
					Message:     "Already placed bet",
					SendAsReply: true,
				}
			}

			if len(currentBet.options) > 1 && !hasOption(currentBet.options, args.Parameters["option"]) {
				return &tmi.OutgoingMessage{
					Channel:     args.Channel,
					ParentID:    args.MsgID,
					Message:     fmt.Sprintf("You need to bet on one of: %s", strings.Join(currentBet.options, ",")),
					SendAsReply: true,
				}
			}

			wagerValue, err := strconv.Atoi(args.Parameters["amount"])
			if err != nil {
				return &tmi.OutgoingMessage{
					Channel:     args.Channel,
					ParentID:    args.MsgID,
					Message:     "Invalid wager point value",
					SendAsReply: true,
				}
			}

			replyChan := make(chan StorageResponse)
			select {
			case gm.storageReqChannel <- StorageRequest{
				Action:      ActionGetPoints,
				ChannelName: args.Channel,
				Username:    args.Username,
				ReplyChan:   replyChan,
			}:
			case <-time.After(time.Second):
				logrus.Error("store request timeout")
				return nil
			}

			var points int64
			select {
			case msg := <-replyChan:
				points = msg.Points
			case <-time.After(time.Second):
				logrus.Error("store request timeout")
				return nil
			}

			if points < int64(wagerValue) {
				if points <= 0 {
					return nil
				}
				wagerValue = int(points)
			}

			select {
			case gm.storageReqChannel <- StorageRequest{
				Action:      ActionSubPoints,
				ChannelName: args.Channel,
				Username:    args.Username,
				Points:      wagerValue,
			}:
			case <-time.After(time.Second):
				logrus.Error("store request timeout")
				return nil
			}

			currentBet.wagers[args.Username] = wager{
				value:  wagerValue,
				option: args.Parameters["option"],
			}

			return nil
		},
	}
}

func GamblingAdminCommand(m tmi.Module) tmi.Command {
	gm := m.(*gamblingMod)
	return tmi.Command{
		Name:                     "betting",
		Description:              "Interact with betting system",
		RequiresBroadcasterOrMod: true,
		Params: []tmi.Parameter{
			{Name: "action", Required: true},
		},
		AllowRestParams: true,
		SubCommands: []tmi.Command{
			{
				Name:        "start",
				Description: "Start a new bet",
				Params: []tmi.Parameter{
					{Name: "timeframe", Required: true},
					{Name: "method", Required: true},
				},
				AllowRestParams: true,
				Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
					return gm.startBet(args)
				},
			},
			{
				Name:        "cancel",
				Description: "Cancels a bet without resolution",
				Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
					return gm.cancelBet(args)
				},
			},
			{
				Name:        "end",
				Description: "Ends a bet with a resultion",
				Params: []tmi.Parameter{
					{Name: "winner", Required: true},
				},
				Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
					return gm.setResult(args)
				},
			},
		},
	}
}

func (gm *gamblingMod) startBet(args tmi.CommandArgs) *tmi.OutgoingMessage {
	_, hasBet := gm.runningBets[args.Channel]
	if hasBet {
		return &tmi.OutgoingMessage{
			Channel: args.Channel,
			Message: "Already running bet NotLikeThis",
		}
	}

	if args.Parameters["timeframe"] == "" {
		return &tmi.OutgoingMessage{
			Channel: args.Channel,
			Message: "Missing timeframe NotLikeThis",
		}
	}

	method := args.Parameters["method"]
	if method != "pool" && method != "multi" {
		return &tmi.OutgoingMessage{
			Channel: args.Channel,
			Message: "Invalid method NotLikeThis",
		}
	}

	secondsToStart, err := strconv.Atoi(args.Parameters["timeframe"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err,
			"timeframe_str": args.Parameters["timeframe"],
		}).Warn("unable to parse timeframe")
	}

	newBet := &gamblingBet{
		isOpen:  true,
		method:  method,
		channel: args.Channel,
		wagers:  make(map[string]wager),
	}

	retMsg := fmt.Sprintf("Bet started. You have %d seconds to place your bet!!!", secondsToStart)

	newBet.isChatterBet = true
	if len(args.RestParams) > 1 {
		newBet.isChatterBet = false
		newBet.options = args.RestParams

		retMsg = fmt.Sprintf("%s Options are: %s", retMsg, strings.Join(args.RestParams, ","))
	}

	go func() {
		<-time.After(time.Second * time.Duration(secondsToStart))
		gm.closingChan <- newBet
	}()

	gm.runningBets[args.Channel] = newBet

	return &tmi.OutgoingMessage{
		Channel: args.Channel,
		Message: retMsg,
	}
}

func (gm *gamblingMod) cancelBet(args tmi.CommandArgs) *tmi.OutgoingMessage {
	currentBet := gm.runningBets[args.Channel]
	if currentBet == nil {
		return &tmi.OutgoingMessage{
			Message: "No bet to cancel SeemsGood",
			Channel: args.Channel,
		}
	}

	delete(gm.runningBets, args.Channel)

	return &tmi.OutgoingMessage{
		Channel: args.Channel,
		Message: "Ok, cancelled!",
	}
}

func (gm *gamblingMod) setResult(args tmi.CommandArgs) *tmi.OutgoingMessage {
	currentBet := gm.runningBets[args.Channel]
	if currentBet == nil {
		return &tmi.OutgoingMessage{
			Message: "No bet to end",
			Channel: args.Channel,
		}
	}

	winnerOptions := args.Parameters["winner"]
	if !currentBet.isChatterBet && !hasOption(currentBet.options, winnerOptions) {
		return &tmi.OutgoingMessage{
			Message: "Invalid winner option",
			Channel: args.Channel,
		}
	}
	if currentBet.isChatterBet {
		winnerOptions = strings.ToLower(winnerOptions)
	}

	if currentBet.method == "pool" {
		return gm.poolMethodResult(currentBet, args, winnerOptions)
	} else {
		return gm.multiMethodResult(currentBet, args, winnerOptions)
	}
}

func (gm *gamblingMod) poolMethodResult(currentBet *gamblingBet, args tmi.CommandArgs, winnerOptions string) *tmi.OutgoingMessage {
	winners := map[string]int{}
	loserPoints := 0
	winnerpoints := 0
	for username, w := range currentBet.wagers {
		isWinner := false
		if currentBet.isChatterBet && winnerOptions == username {
			isWinner = true
		}
		if !currentBet.isChatterBet && w.option == winnerOptions {
			isWinner = true
		}

		if isWinner {
			winners[username] = w.value
			winnerpoints += w.value
		} else {
			loserPoints += w.value
		}
	}

	if len(winners) <= 0 {
		return &tmi.OutgoingMessage{
			Channel: args.Channel,
			Message: "No winners. House wins LUL",
		}
	}

	for username, basePoints := range winners {
		userWinnings := (currentBet.wagers[username].value / winnerpoints) * loserPoints
		select {
		case gm.storageReqChannel <- StorageRequest{
			Action:      ActionAddPoints,
			ChannelName: args.Channel,
			Username:    username,
			Points:      basePoints + userWinnings,
		}:
		case <-time.After(time.Second * 2):
			logrus.Error("store request timeout")
		}

		gm.client.Send(&tmi.OutgoingMessage{
			Channel: args.Channel,
			Message: fmt.Sprintf("%s won %d points", username, basePoints+userWinnings),
		})
	}

	return nil
}

func (gm *gamblingMod) multiMethodResult(currentBet *gamblingBet, args tmi.CommandArgs, winnerOptions string) *tmi.OutgoingMessage {
	for username, w := range currentBet.wagers {
		isWinner := false
		if currentBet.isChatterBet && winnerOptions == username {
			isWinner = true
		}
		if !currentBet.isChatterBet && w.option == winnerOptions {
			isWinner = true
		}

		if isWinner {
			select {
			case gm.storageReqChannel <- StorageRequest{
				Action:      ActionAddPoints,
				ChannelName: args.Channel,
				Username:    username,
				Points:      w.value * 2,
			}:
			case <-time.After(time.Second * 2):
				logrus.Error("store request timeout")
			}

			gm.client.Send(&tmi.OutgoingMessage{
				Channel: args.Channel,
				Message: fmt.Sprintf("%s won %d points", username, w.value*10),
			})
		}
	}

	return nil
}
