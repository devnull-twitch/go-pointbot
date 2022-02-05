package pointbot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/devnull-twitch/go-tmi"
	"github.com/sirupsen/logrus"
)

func PPCConfigModuleCommand() tmi.ModuleCommand {
	return tmi.ModuleCommand{
		Command: tmi.Command{
			Name:        "ppc",
			Description: "Allows mods to set the points per chat",
			Params: []tmi.Parameter{
				{
					Name:     "value",
					Required: true,
				},
			},
			RequiresBroadcasterOrMod: true,
		},
		ModuleCommandHandler: func(client *tmi.Client, m tmi.Module, args tmi.CommandArgs) *tmi.OutgoingMessage {
			points, err := strconv.Atoi(args.Parameters["value"])
			if err != nil {
				return &tmi.OutgoingMessage{
					Message: "Unable to read new ppc NotLikeThis",
				}
			}

			pm := m.(*pointModule)
			pm.storageReqChannel <- StorageRequest{
				Action:      ActionSetPPC,
				ChannelName: args.Channel,
				Points:      points,
			}
			return nil
		},
	}
}

func (s *Storage) SetChannelPPC(channelID int64, ppc int) error {
	_, err := s.conn.Exec(context.Background(), "UPDATE channels SET points_per_chat = $2 WHERE id = $1", channelID, ppc)
	if err != nil {
		return fmt.Errorf("unable to set new ppc: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"cid": channelID,
		"ppc": ppc,
	}).Info("set new ppc")
	return nil
}
