package api

import (
	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-tmi"
	"github.com/gin-gonic/gin"
)

func Run(bot *tmi.Client, storageReqChannel chan<- pointbot.StorageRequest) {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.POST("/bot/join/:channel", GetJoinHandler(bot, storageReqChannel))
	r.GET("/bot/points/:token", GetTopPointHandler(storageReqChannel))
	r.POST("/bot/points/:token/:chatter/:points", GetAddPointHandler(storageReqChannel))

	r.Run(":8085")
}
