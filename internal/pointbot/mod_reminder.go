package pointbot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/devnull-twitch/go-tmi"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nicklaw5/helix"
	"github.com/sirupsen/logrus"
)

type reminderMod struct {
	client      *tmi.Client
	conn        *pgxpool.Pool
	twClient    *helix.Client
	hadMessages map[string]bool
}

func ReminderMod(client *tmi.Client, conn *pgxpool.Pool, twClient *helix.Client) tmi.Module {
	mod := &reminderMod{
		client:      client,
		conn:        conn,
		twClient:    twClient,
		hadMessages: make(map[string]bool),
	}
	return mod
}

func (*reminderMod) ExternalTrigger(_ *tmi.Client) <-chan *tmi.ModuleArgs {
	moduleTrigger := make(chan *tmi.ModuleArgs)
	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			moduleTrigger <- &tmi.ModuleArgs{}
		}
	}()
	return moduleTrigger
}

func (rm *reminderMod) Handler(client *tmi.Client, _ tmi.ModuleArgs) *tmi.OutgoingMessage {
	channelIsLive := map[string]bool{}
	rows, err := rm.conn.Query(context.Background(), `
		SELECT c.channel_name, rm.reminder_message
		FROM reminder_messages as rm
		JOIN channels AS c ON rm.channel_id = c.id
		WHERE rm.last_send < current_timestamp - rm.interval`)

	if err != nil {
		logrus.WithError(err).Error("database error")
		return nil
	}

	channelsMessagesSend := make([]string, 0)

	for rows.Next() {
		var (
			channelname string
			message     string
		)
		if err := rows.Scan(&channelname, &message); err != nil {
			logrus.WithError(err).Error("error scanning database result")
			return nil
		}

		if !rm.hadMessages[channelname] {
			return nil
		}

		isLive, known := channelIsLive[channelname]
		if !known {
			userListResp, err := rm.twClient.GetUsers(&helix.UsersParams{
				Logins: []string{channelname},
			})
			if err != nil {
				logrus.WithError(err).Error("unable to load user")
				continue
			}

			if len(userListResp.Data.Users) <= 0 {
				logrus.Error("user not found")
				continue
			}

			streamResp, err := rm.twClient.GetStreams(&helix.StreamsParams{
				UserIDs: []string{userListResp.Data.Users[0].ID},
			})
			if err != nil {
				logrus.WithError(err).Error("unable to load streams")
				continue
			}

			if len(streamResp.Data.Streams) <= 0 && streamResp.Data.Streams[0].Type != "live" {
				channelIsLive[channelname] = false
				continue
			}

			channelIsLive[channelname] = true
			isLive = true
		}

		if isLive {
			channelsMessagesSend = append(channelsMessagesSend, channelname)
			client.Send(&tmi.OutgoingMessage{
				Message: message,
				Channel: channelname,
			})
		}
	}

	for _, channelname := range channelsMessagesSend {
		rm.hadMessages[channelname] = false
	}

	_, err = rm.conn.Exec(context.Background(), "DELETE FROM reminder_messages WHERE last_send < current_timestamp - interval AND trigger_once = TRUE")
	if err != nil {
		logrus.WithError(err).Error("unable to delte one time reminders")
	}

	_, err = rm.conn.Exec(context.Background(), "UPDATE reminder_messages SET last_send = current_timestamp WHERE last_send < current_timestamp - interval")
	if err != nil {
		logrus.WithError(err).Error("unable to update timestamp of reminder messages")
	}

	return nil
}

func (rm *reminderMod) MessageTrigger(_ *tmi.Client, incoming *tmi.IncomingMessage) *tmi.ModuleArgs {
	if os.Getenv("USERNAME") != incoming.Username {
		rm.hadMessages[incoming.Channel] = true
	}

	return nil
}

func ReminderCommand(m tmi.Module) tmi.Command {
	rm := m.(*reminderMod)
	return tmi.Command{
		Name:                     "reminder",
		Description:              "Configure reminders",
		RequiresBroadcasterOrMod: true,
		SubCommands: []tmi.Command{
			{
				Name:        "add",
				Description: "Add a new reminder message",
				Params: []tmi.Parameter{
					{Name: "time", Required: true},
					{Name: "message", Required: true},
					{Name: "triggerOnce", Default: strptr("")},
				},
				AllowRestParams:          false,
				RequiresBroadcasterOrMod: true,
				Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
					triggerOnce := false
					if args.Parameters["triggerOnce"] != "" {
						triggerOnce = true
					}

					interval, err := time.ParseDuration(args.Parameters["time"])
					if err != nil {
						logrus.WithError(err).Warn("duration parser error")
						return &tmi.OutgoingMessage{
							Channel: args.Channel,
							Message: "Cannot understand the interval time",
						}
					}

					if err := rm.addReminderMessage(
						args.Channel,
						args.Parameters["message"],
						interval,
						triggerOnce,
					); err != nil {
						logrus.WithError(err).Warn("reminder save error")
						return &tmi.OutgoingMessage{
							Channel: args.Channel,
							Message: "Unable to save reminder NotLikeThis",
						}
					}

					return &tmi.OutgoingMessage{
						Channel: args.Channel,
						Message: "Ok SeemsGood",
					}
				},
			},
			{
				Name:                     "remove",
				Description:              "Removes a reminder message",
				RequiresBroadcasterOrMod: true,
				Params: []tmi.Parameter{
					{Name: "message", Required: true},
				},
				Handler: func(client *tmi.Client, args tmi.CommandArgs) *tmi.OutgoingMessage {
					if err := rm.removeReminderMessage(args.Channel, args.Parameters["message"]); err != nil {
						logrus.WithError(err).Warn("reminder deletion error")
						return &tmi.OutgoingMessage{
							Channel: args.Channel,
							Message: "Unable to remove that NotLikeThis",
						}
					}

					return &tmi.OutgoingMessage{
						Channel: args.Channel,
						Message: "Ok SeemsGood",
					}
				},
			},
		},
	}
}

func (rm *reminderMod) addReminderMessage(
	channelName string,
	message string,
	interval time.Duration,
	triggerOnce bool,
) error {
	channelID, err := getChannelIdByName(rm.conn, channelName)
	if err != nil {
		return err
	}

	sqlInverval := pgtype.Interval{}
	if err := sqlInverval.Set(interval); err != nil {
		return fmt.Errorf("unable to set duration as interval: %w", err)
	}

	_, err = rm.conn.Exec(
		context.Background(),
		`INSERT INTO reminder_messages 
			(channel_id, reminder_message, interval, trigger_once, last_send)
		VALUES
			( $1, $2, $3, $4, current_timestamp )`,
		channelID,
		message,
		sqlInverval,
		triggerOnce,
	)
	if err != nil {
		return err
	}

	return nil
}

func (rm *reminderMod) removeReminderMessage(channelName, message string) error {
	channelID, err := getChannelIdByName(rm.conn, channelName)
	if err != nil {
		return err
	}

	_, err = rm.conn.Exec(
		context.Background(),
		`DELETE FROM reminder_messages WHERE reminder_message = $1 AND channel_id = $2`,
		message,
		channelID,
	)
	if err != nil {
		return err
	}

	return nil
}

func strptr(s string) *string {
	return &s
}
