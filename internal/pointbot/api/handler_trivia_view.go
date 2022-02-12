package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

func GetTrivaRender(conn *pgxpool.Pool) gin.HandlerFunc {
	type TriviaRenderObj struct {
		Message          string
		ProviderUsername string
		Question         string
		Channel          string
		Answer1          string
		Answer2          string
		Answer3          string
		Answer4          string
	}
	return func(c *gin.Context) {
		cid, err := getChannelIdByName(conn, c.Param("channel"))
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		row := conn.QueryRow(
			context.Background(),
			`SELECT 
				username, question, correct_answer, wrong_answer_1, wrong_answer_2, wrong_answer_3
			FROM trivia_questions
			WHERE 
				channel_id = $1 AND
				active = 1 AND
				wrong_answer_1 IS NOT NULL AND
				wrong_answer_2 IS NOT NULL AND
				wrong_answer_3 IS NOT NULL`,
			cid,
		)

		renderObj := &TriviaRenderObj{
			Message:          "",
			ProviderUsername: "",
			Question:         "Load me",
			Channel:          "iamdevnull",
			Answer1:          "1",
			Answer2:          "2",
			Answer3:          "3",
			Answer4:          "4",
		}

		err = row.Scan(
			&renderObj.ProviderUsername,
			&renderObj.Question,
			&renderObj.Answer1,
			&renderObj.Answer2,
			&renderObj.Answer3,
			&renderObj.Answer4,
		)
		if err != nil {
			logrus.WithError(err).Error("unable to load question from DB")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.HTML(http.StatusOK, "trivia.tmpl", renderObj)
	}
}

func getChannelIdByName(conn *pgxpool.Pool, channelName string) (int64, error) {
	channelrow := conn.QueryRow(context.Background(), "SELECT id FROM channels WHERE channel_name = $1", channelName)
	var channelID int64
	if err := channelrow.Scan(&channelID); err != nil {
		return 0, fmt.Errorf("unable to load channel ID: %w", err)
	}

	return channelID, nil
}
