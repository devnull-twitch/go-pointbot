package overlay

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func Setup(r gin.IRouter, conn *pgxpool.Pool) {
	r.Static("/overlay/assets", os.Getenv("OVERLAY_ASSET_DIR"))

	r.GET("/overlay/view", func(c *gin.Context) {
		type (
			renderLeaderboardPosition struct {
				Name   string
				Points int
			}
			renderLeaderboard struct {
				Positions []renderLeaderboardPosition
			}
			renderMusicTrack struct {
				Song   string
				Artist string
			}
			renderiFrame struct {
				URL string
			}
			renderBoxes struct {
				Left         int
				Right        int
				Top          int
				Bottom       int
				BgColor      string
				Leaderboard  *renderLeaderboard
				TopPoints    *renderLeaderboardPosition
				CurrentTrack *renderMusicTrack
				IFrame       *renderiFrame
			}
			renderParamsType struct {
				ChannelName string
				EditMode    bool
				Boxes       []*renderBoxes
			}
		)
		renderParams := &renderParamsType{
			EditMode:    true,
			ChannelName: "Test",
			Boxes: []*renderBoxes{
				{
					Right:   50,
					Top:     90,
					BgColor: "#7DD3FC",
					Leaderboard: &renderLeaderboard{Positions: []renderLeaderboardPosition{
						{Name: "Demo", Points: 100},
						{Name: "Twitch viewer", Points: 80},
					}},
				},
				{
					Left:    30,
					Top:     30,
					BgColor: "#7DD3FC",
					IFrame: &renderiFrame{
						URL: "https://tailwindcss.com/docs/top-right-bottom-left",
					},
				},
				{
					Right:   230,
					Top:     30,
					BgColor: "#7DD3FC",
					CurrentTrack: &renderMusicTrack{
						Song:   "Song",
						Artist: "Songwriter",
					},
				},
				{
					Right:   30,
					Top:     230,
					BgColor: "#7DD3FC",
					TopPoints: &renderLeaderboardPosition{
						Name:   "theCodeBug",
						Points: 1000,
					},
				},
			},
		}
		c.HTML(http.StatusOK, "overlay.html", renderParams)
	})
	r.GET("/overlay/edit")

	r.POST("/overlay/box/:boxname")
	r.PUT("/overlay/box/:boxname")
	r.DELETE("/overlay/box/:boxname")
}
