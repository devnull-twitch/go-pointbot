package api

import (
	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-tmi"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func Setup(
	r gin.IRouter,
	bot *tmi.Client,
	storageReqChannel chan<- pointbot.StorageRequest,
	conn *pgxpool.Pool,
) {
	r.GET("/points/:token", GetTopPointHandler(storageReqChannel))
	r.POST("/points/:token/:chatter/:points", GetAddPointHandler(storageReqChannel))

	r.POST("/track/:token", GetTrackPostHandler(conn))
	r.GET("/track/:token", GetTrackGetHanmdler(conn))
}
