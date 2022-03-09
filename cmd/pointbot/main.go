package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-pointbot/internal/pointbot/api"
	"github.com/devnull-twitch/go-tmi"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix"
)

func main() {
	godotenv.Load(".env.yaml")

	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to database: %w", err))
	}

	bot, err := tmi.New(os.Getenv("USERNAME"), os.Getenv("TOKEN"), os.Getenv("COMMAND_MARK"))
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to IRC: %w", err))
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:       os.Getenv("TW_CLIENTID"),
		AppAccessToken: os.Getenv("TW_APP_ACCESS"),
	})
	if err != nil {
		log.Fatal("unable to create twitch api client")
	}

	bot.AddCommand(pointbot.ShoutoutCommand(client))

	storageReqChannel := pointbot.NewStorage(conn)

	pm := pointbot.PointModule(storageReqChannel)
	bot.AddModule(pm)
	bot.AddCommand(pointbot.PointModuleCommand(pm))
	bot.AddCommand(pointbot.PPCConfigModuleCommand(pm))
	bot.AddCommand(pointbot.LeaderboardCommand(pm))

	triviaStoreReqChan := pointbot.NewTriviaStorage(conn)
	tm := pointbot.TriviaModule(triviaStoreReqChan)
	bot.AddModule(tm)
	bot.AddCommand(pointbot.TriviaCommand(tm))

	gambling := pointbot.GamblingModule(bot, storageReqChannel)
	bot.AddModule(gambling)
	bot.AddCommand(pointbot.GamblingAdminCommand(gambling))
	bot.AddCommand(pointbot.GamblingUserCommand(gambling))

	bot.AddCommand(pointbot.StopwatchCommand())

	reminderModule := pointbot.ReminderMod(bot, conn, client)
	bot.AddModule(reminderModule)
	bot.AddCommand(pointbot.ReminderCommand(reminderModule))

	bot.AddCommand(pointbot.DemoCmd())
	bot.AfterStartup(func() {
		channelReply := make(chan pointbot.StorageResponse)
		storageReqChannel <- pointbot.StorageRequest{
			Action:    pointbot.ActionChannels,
			ReplyChan: channelReply,
		}

		for r := range channelReply {
			bot.JoinChannel(r.ChannelName)
		}
	})

	go api.Run(bot, storageReqChannel, conn)

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
