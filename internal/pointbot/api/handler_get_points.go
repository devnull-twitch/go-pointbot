package api

import (
	"net/http"
	"time"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetTopPointHandler(storageReqChannel chan<- pointbot.StorageRequest) gin.HandlerFunc {
	type apiResponse struct {
		Username string `json:"user"`
		Points   int    `json:"points"`
	}
	return func(c *gin.Context) {
		replyChan := make(chan pointbot.StorageResponse)

		select {
		case storageReqChannel <- pointbot.StorageRequest{
			Action:    pointbot.ActionList,
			Token:     c.Param("token"),
			ReplyChan: replyChan,
		}:
		case <-time.After(time.Second):
			logrus.Warn("Storage didnt read request")
			c.Status(http.StatusInternalServerError)
			return
		}

		resps := make([]apiResponse, 0)
		goon := true
		for goon {
			select {
			case r, ok := <-replyChan:
				if !ok {
					goon = false
					break
				}
				if r.Username == "" {
					continue
				}

				resps = append(resps, apiResponse{
					Username: r.Username,
					Points:   r.Points,
				})
			case <-time.After(time.Second):
				goon = false
				if len(resps) <= 0 {
					logrus.Warn("Storage didnt send and more data?")
					c.Status(http.StatusInternalServerError)
					return
				}
			}
		}

		c.JSON(http.StatusOK, resps)
	}
}
