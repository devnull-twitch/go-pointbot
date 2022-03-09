package pointbot

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type FeatureFlagChecker interface {
	FeatureFlagAccptanceCheck(*tmi.IncomingCommand) bool
}

type (
	dbFeatureFlagChecker struct {
		conn    *pgxpool.Pool
		reqChan chan<- CheckRequest
	}

	CheckRequest struct {
		ChannelID   int
		ChannelName string
		Command     string
		ReplyChan   chan bool
	}
)

func NewFeatureFlagChecker(conn *pgxpool.Pool) FeatureFlagChecker {
	ffc := &dbFeatureFlagChecker{conn: conn}
	reqChan := ffc.Worker()
	ffc.reqChan = reqChan

	return ffc
}

func (ffc *dbFeatureFlagChecker) FeatureFlagAccptanceCheck(incoming *tmi.IncomingCommand) bool {
	replyChan := make(chan bool)
	select {
	case ffc.reqChan <- CheckRequest{
		ChannelName: incoming.Channel,
		Command:     incoming.Command,
		ReplyChan:   replyChan,
	}:
	case <-time.After(time.Second):
		logrus.Error("feature flag request timout")
		return true
	}

	select {
	case resp := <-replyChan:
		return resp
	case <-time.After(time.Second):
		logrus.Error("feature flag response timout")
		return true
	}
}

func (ffc *dbFeatureFlagChecker) Worker() chan CheckRequest {
	reqChan := make(chan CheckRequest)
	go func() {
		for {
			r := <-reqChan
			if r.ChannelID == 0 && r.ChannelName != "" {
				id, err := getChannelIdByName(ffc.conn, r.ChannelName)
				if err != nil {
					logrus.WithError(err).Error("unable to load channel id for feature flag check")
					replyWith(r.ReplyChan, true)
					return
				}

				r.ChannelID = int(id)
			}
			row := ffc.conn.QueryRow(
				context.Background(),
				"SELECT flag_value FROM feature_flags WHERE channel_id = $1 and command = $2",
				r.ChannelID,
				r.Command,
			)
			var flagValue bool
			if err := row.Scan(&flagValue); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					replyWith(r.ReplyChan, true)
					return
				}
				logrus.WithError(err).Error("error loading feature flag")
				replyWith(r.ReplyChan, false)
			}

			replyWith(r.ReplyChan, flagValue)
		}
	}()

	return reqChan
}

func replyWith(replyChan chan bool, fv bool) {
	select {
	case replyChan <- fv:
	case <-time.After(time.Second):
	}
}

func FeatureFlagAdminCommand(conn *pgxpool.Pool) tmi.Command {
	return tmi.Command{
		Name:        "ffc",
		Description: "Enable er disable bot commands",
		Params: []tmi.Parameter{
			{Name: "command", Required: true},
			{Name: "flag", Required: true, Validate: func(s string) bool {
				num, err := strconv.Atoi(s)
				return err == nil && (num == 1 || num == 0)
			}},
		},
		RequiresBroadcasterOrMod: true,
		Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
			channelID, err := getChannelIdByName(conn, args.Channel)
			if err != nil {
				logrus.WithError(err).Error("unable to load channel id")
				return &tmi.OutgoingMessage{
					Channel: args.Channel,
					Message: "Unable to save that",
				}
			}

			num, _ := strconv.Atoi(args.Parameters["flag"])

			_, err = conn.Exec(
				context.Background(),
				`INSERT INTO feature_flags ( channel_id, command, flag_value ) VALUES ( $1, $2, $3 )
				ON CONFLICT ON CONSTRAINT unique_command_in_channel
				DO UPDATE SET flag_value = $3`,
				channelID,
				args.Parameters["command"],
				num == 1,
			)
			if err != nil {
				logrus.WithError(err).Error("unable to set feature flag")
				return &tmi.OutgoingMessage{
					Channel: args.Channel,
					Message: "Unable to save that",
				}
			}

			return &tmi.OutgoingMessage{
				Channel: args.Channel,
				Message: "Ok",
			}
		},
	}
}
