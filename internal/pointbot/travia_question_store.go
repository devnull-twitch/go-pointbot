package pointbot

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type (
	TriviaQuestionReq struct {
		ChannelName   string
		Username      string
		Question      string
		CorrectAnswer string
		Wrong1        string
		Wrong2        string
		Wrong3        string
	}
)

func NewTriviaStorage(conn *pgxpool.Pool) chan<- TriviaQuestionReq {
	requestChan := make(chan TriviaQuestionReq)
	go func(reqs <-chan TriviaQuestionReq) {
		for {
			req := <-reqs

			var (
				wrong1 sql.NullString = sql.NullString{}
				wrong2 sql.NullString = sql.NullString{}
				wrong3 sql.NullString = sql.NullString{}
			)

			channelID, err := getChannelIdByName(conn, req.ChannelName)
			if err != nil {
				logrus.WithError(err).Error("unable to get channel id")
				return
			}

			if req.Wrong1 != "" {
				wrong1 = sql.NullString{Valid: true, String: req.Wrong1}
			}
			if req.Wrong2 != "" {
				wrong2 = sql.NullString{Valid: true, String: req.Wrong2}
			}
			if req.Wrong3 != "" {
				wrong3 = sql.NullString{Valid: true, String: req.Wrong3}
			}

			_, err = conn.Exec(
				context.Background(),
				`INSERT INTO trivia_questions 
					( channel_id, username, question, correct_answer, wrong_answer_1, wrong_answer_2, wrong_answer_3) 
					VALUES
					( $1, $2, $3, $4, $5, $6, $7 )`,
				channelID,
				req.Username,
				req.Question,
				req.CorrectAnswer,
				wrong1,
				wrong2,
				wrong3,
			)

			if err != nil {
				logrus.WithError(err).Error("unable to save new trivia question")
			}
		}
	}(requestChan)

	return requestChan
}

func getChannelIdByName(conn *pgxpool.Pool, channelName string) (int64, error) {
	channelrow := conn.QueryRow(context.Background(), "SELECT id FROM channels WHERE channel_name = $1", channelName)
	var channelID int64
	if err := channelrow.Scan(&channelID); err != nil {
		return 0, fmt.Errorf("unable to load channel ID: %w", err)
	}

	return channelID, nil
}
