package main

import (
	"log"
	"os"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-pointbot/internal/pointbot/api"
	"github.com/devnull-twitch/go-tmi"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.yaml")

	bot, err := tmi.New(os.Getenv("USERNAME"), os.Getenv("TOKEN"), os.Getenv("COMMAND_MARK"))
	if err != nil {
		log.Fatal(err)
	}

	storageReqChannel := pointbot.NewStorage()

	bot.AddCommand(pointbot.DemoCmd())
	bot.AddCommand(pointbot.ShowPoints(storageReqChannel))

	go api.Run(bot, storageReqChannel)

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
