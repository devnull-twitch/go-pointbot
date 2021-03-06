package api

import (
	"time"

	"github.com/devnull-twitch/go-pointbot/internal/pointbot"
	"github.com/devnull-twitch/go-tmi"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func Run(bot *tmi.Client, storageReqChannel chan<- pointbot.StorageRequest, conn *pgxpool.Pool) {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowWildcard:    true,
		AllowMethods:     []string{"PUT", "POST", "GET"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	r.POST("/bot/join/:channel", GetJoinHandler(bot, storageReqChannel))
	r.GET("/bot/points/:token", GetTopPointHandler(storageReqChannel))
	r.POST("/bot/points/:token/:chatter/:points", GetAddPointHandler(storageReqChannel))

	r.POST("/bot/track/:token", GetTrackPostHandler(conn))
	r.GET("/bot/track/:token", GetTrackGetHanmdler(conn))

	r.Run(":8085")
}
