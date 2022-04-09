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

func (*trviaModule) ExternalTrigger(_ *tmi.Client) <-chan *tmi.ModuleArgs {
	return nil
}

func (*trviaModule) MessageTrigger(_ *tmi.Client, _ *tmi.IncomingMessage) *tmi.ModuleArgs {
	return nil
}

func (*trviaModule) Handler(_ *tmi.Client, _ tmi.ModuleArgs) *tmi.OutgoingMessage {
	return nil
}

func TriviaCommand(m tmi.Module) tmi.Command {
	emptyStr := ""
	tm := m.(*trviaModule)
	return tmi.Command{
		Name:        "trivia",
		Description: "Add trivia questions",
		Params: []tmi.Parameter{
			{Name: "question", Required: true},
			{Name: "correct", Required: true},
			{Name: "wrong1", Default: &emptyStr},
			{Name: "wrong2", Default: &emptyStr},
			{Name: "wrong3", Default: &emptyStr},
		},
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
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
	}
}
