package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-pointbot/internal/pointbot/api"
	"github.com/devnull-twitch/go-tmi"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.yaml")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to database: %w", err))
	}
	defer conn.Close(context.Background())

	bot, err := tmi.New(os.Getenv("USERNAME"), os.Getenv("TOKEN"), os.Getenv("COMMAND_MARK"))
	if err != nil {
		log.Fatal(fmt.Errorf("unable to connect to IRC: %w", err))
	}

	storageReqChannel := pointbot.NewStorage(conn)

	pm := pointbot.PointModule(storageReqChannel)
	bot.AddModule(pm)
	bot.AddModuleCommand(pm, pointbot.PointModuleCommand())

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

	go api.Run(bot, storageReqChannel)

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
