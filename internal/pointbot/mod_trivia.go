package pointbot

import (
	"github.com/devnull-twitch/go-tmi"
)

type trviaModule struct {
	storage chan<- TriviaQuestionReq
}

func TriviaModule(storageReqChannel chan<- TriviaQuestionReq) tmi.Module {
	return &trviaModule{
		storage: storageReqChannel,
	}
}

func (tm *trviaModule) ExternalTrigger(client *tmi.Client) <-chan *tmi.ModuleArgs {
	return nil
}

func (tm *trviaModule) MessageTrigger(client *tmi.Client, incoming *tmi.IncomingMessage) *tmi.ModuleArgs {
	return nil
}

func (tm *trviaModule) Handler(client *tmi.Client, args tmi.ModuleArgs) *tmi.OutgoingMessage {
	return nil
}

func TriviaCommand() tmi.ModuleCommand {
	emptyStr := ""
	return tmi.ModuleCommand{
		ModuleCommandHandler: func(client *tmi.Client, m tmi.Module, args tmi.CommandArgs) *tmi.OutgoingMessage {
			tm := m.(*trviaModule)
			tm.storage <- TriviaQuestionReq{
				ChannelName:   args.Channel,
				Username:      args.Username,
				Question:      args.Parameters["question"],
				CorrectAnswer: args.Parameters["correct"],
				Wrong1:        args.Parameters["wrong1"],
				Wrong2:        args.Parameters["wrong2"],
				Wrong3:        args.Parameters["wrong3"],
			}
			return &tmi.OutgoingMessage{
				Message: "Thanks SeemsGood",
			}
		},
		Command: tmi.Command{
			Name:        "trivia",
			Description: "Add trivia questions",
			Params: []tmi.Parameter{
				{Name: "question", Required: true},
				{Name: "correct", Required: true},
				{Name: "wrong1", Default: &emptyStr},
				{Name: "wrong2", Default: &emptyStr},
				{Name: "wrong3", Default: &emptyStr},
			},
		},
	}
}
