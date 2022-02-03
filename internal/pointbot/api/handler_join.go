package api

import (
	"net/http"
	"time"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-tmi"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetJoinHandler(bot *tmi.Client, storageReqChannel chan<- pointbot.StorageRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelToken := String(20)
		channelName := c.Param("channel")
		select {
		case storageReqChannel <- pointbot.StorageRequest{Action: pointbot.ActionJoin, ChannelName: channelName, Token: channelToken}:
		case <-time.After(time.Second):
			logrus.Warn("Storage didnt read request")
		}

		bot.JoinChannel(channelName)

		c.String(http.StatusOK, channelToken)
	}
}
