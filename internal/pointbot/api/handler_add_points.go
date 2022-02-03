package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetAddPointHandler(storageReqChannel chan<- pointbot.StorageRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		intPoints, err := strconv.Atoi(c.Param("points"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		select {
		case storageReqChannel <- pointbot.StorageRequest{
			Action:   pointbot.ActionAddPoints,
			Token:    c.Param("token"),
			Username: c.Param("chatter"),
			Points:   intPoints,
		}:
		case <-time.After(time.Second):
			logrus.Warn("Storage didnt read request")
		}

		c.Status(http.StatusCreated)
	}
}
