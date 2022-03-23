package pointbot

import (
	"fmt"

	"github.com/devnull-twitch/go-tmi"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

func ShoutoutCommand(twClient *helix.Client, ffc FeatureFlagChecker) tmi.Command {
	return tmi.Command{
		Name:        "so",
		Description: "Shoutout other streamer",
		Params: []tmi.Parameter{
			{Name: "username", Required: true},
		},
		RequiresBroadcasterOrMod: true,
		AllowRestParams:          false,
		AcceptanceCheck:          ffc.FeatureFlagAccptanceCheck,
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			userListResp, err := twClient.GetUsers(&helix.UsersParams{
				Logins: []string{args.Parameters["username"]},
			})
			if err != nil {
				logrus.WithError(err).Error("unable to load user for shoutout")
				return nil
			}

			if len(userListResp.Data.Users) <= 0 {
				logrus.Error("user not found")
				return nil
			}

			channelResp, err := twClient.GetChannelInformation(&helix.GetChannelInformationParams{
				BroadcasterIDs: []string{
					userListResp.Data.Users[0].ID,
				},
			})
			if err != nil {
				logrus.WithError(err).Error("unable to load user for shoutout")
				return nil
			}

			if len(channelResp.Data.Channels) <= 0 {
				logrus.Error("user not found")
				return nil
			}

			channel := channelResp.Data.Channels[0]
			return &tmi.OutgoingMessage{
				Message: fmt.Sprintf(
					"Check Out %s! Last streamed \"%s\" in %s - https://www.twitch.tv/%s",
					channel.BroadcasterName,
					channel.Title,
					channel.GameName,
					channel.BroadcasterName,
				),
			}
		},
	}
}
