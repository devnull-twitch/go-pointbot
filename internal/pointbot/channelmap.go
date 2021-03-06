package pointbot

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type (
	Action          int
	StorageResponse struct {
		ChannelName string
		Username    string
		Points      int64
	}
	StorageRequest struct {
		Action      Action
		ChannelName string
		Token       string
		Username    string
		Subscriber  bool
		Points      int
		ReplyChan   chan StorageResponse
	}
	Storage struct {
		conn *pgxpool.Pool
	}
)

const (
	ActionJoin Action = iota + 1
	ActionGetPoints
	ActionAddPoints
	ActionDelPoints
	ActionSubPoints
	ActionTop
	ActionList
	ActionChannels
	ActionChatBasePoint
	ActionSetPPC
)

func NewStorage(conn *pgxpool.Pool) chan<- StorageRequest {
	requestChan := make(chan StorageRequest)
	go func(reqs <-chan StorageRequest) {
		s := &Storage{
			conn: conn,
		}

		for {
			req := <-reqs
			logrus.WithFields(logrus.Fields{
				"action":  req.Action,
				"channel": fmt.Sprintf("%s %s", req.ChannelName, req.Token),
			}).Info("storage request fetched")

			switch req.Action {
			case ActionJoin:
				s.AddChannel(req.Token, req.ChannelName)

			case ActionChannels:
				resps := s.ListChannels()
				if len(resps) <= 0 {
					close(req.ReplyChan)
					continue
				}

				for _, r := range resps {
					select {
					case req.ReplyChan <- r:
					case <-time.After(time.Second):
						logrus.Warn("channel didnt take response")
						continue
					}
				}

				close(req.ReplyChan)

			case ActionAddPoints:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				if err := s.AddPoints(cid, req.Username, req.Points); err != nil {
					logrus.WithError(err).Error("unable to add points")
				}

			case ActionDelPoints:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				if err := s.DeletePoints(cid, req.Username); err != nil {
					logrus.WithError(err).Error("unable to delete points")
				}

			case ActionSubPoints:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				if err := s.AddPoints(cid, req.Username, -req.Points); err != nil {
					logrus.WithError(err).Error("unable to subtract points")
				}

			case ActionGetPoints:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				points := s.GetPoints(cid, req.Username)
				select {
				case req.ReplyChan <- StorageResponse{Username: req.Username, ChannelName: req.ChannelName, Points: points}:
				case <-time.After(time.Second):
					logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
				}

			case ActionTop:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				resps := s.ListPoints(cid, 1)
				if len(resps) <= 0 {
					req.ReplyChan <- StorageResponse{}
					continue
				}

				select {
				case req.ReplyChan <- resps[0]:
				case <-time.After(time.Second):
					logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
				}

			case ActionList:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				resps := s.ListPoints(cid, 10)
				if len(resps) <= 0 {
					close(req.ReplyChan)
					continue
				}

				for _, r := range resps {
					select {
					case req.ReplyChan <- r:
					case <-time.After(time.Second):
						logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
						continue
					}
				}

				close(req.ReplyChan)

			case ActionChatBasePoint:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				if err := s.AddPPC(cid, req.Username, req.Subscriber); err != nil {
					logrus.WithError(err).Error("unable to add PPC")
				}

			case ActionSetPPC:
				cid, ok := s.getChannelId(req)
				if !ok {
					continue
				}
				if err := s.SetChannelPPC(cid, req.Points); err != nil {
					logrus.WithError(err).Error("unable to set PPC")
				}
			}
		}
	}(requestChan)

	return requestChan
}

func (s *Storage) getChannelIdByToken(token string) (int64, error) {
	channelrow := s.conn.QueryRow(context.Background(), "SELECT id FROM channels WHERE token = $1", token)
	var channelID int64
	if err := channelrow.Scan(&channelID); err != nil {
		return 0, fmt.Errorf("unable to load channel ID: %w", err)
	}

	return channelID, nil
}

func (s *Storage) getChannelIdByName(channelName string) (int64, error) {
	channelrow := s.conn.QueryRow(context.Background(), "SELECT id FROM channels WHERE channel_name = $1", channelName)
	var channelID int64
	if err := channelrow.Scan(&channelID); err != nil {
		return 0, fmt.Errorf("unable to load channel ID: %w", err)
	}

	return channelID, nil
}

func (s *Storage) getChannelId(req StorageRequest) (int64, bool) {
	if req.Token == "" && req.ChannelName != "" {
		cid, err := s.getChannelIdByName(req.ChannelName)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"name":  req.ChannelName,
				"token": req.Token,
			}).Warn("unable to load channel ID")
			return 0, false
		}

		return cid, true
	}

	cid, err := s.getChannelIdByToken(req.Token)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"name":  req.ChannelName,
			"token": req.Token,
		}).Warn("unable to load channel ID")
		return 0, false
	}

	return cid, true
}

func (s *Storage) AddChannel(token, channel string) {
	_, err := s.conn.Exec(context.Background(), "INSERT INTO channels (channel_name, token) VALUES ($1, $2)", channel, token)
	if err != nil {
		logrus.WithError(err).Error("unable to add channel to list")
	}
}

func (s *Storage) ListChannels() []StorageResponse {
	channelRows, err := s.conn.Query(context.Background(), "SELECT channel_name FROM channels")
	if err != nil {
		logrus.WithError(err).Error("unable to load channels")
		return []StorageResponse{}
	}

	response := make([]StorageResponse, 0, 50)
	for channelRows.Next() {
		var channelName string
		channelRows.Scan(&channelName)
		response = append(response, StorageResponse{
			ChannelName: channelName,
		})
	}

	return response
}

func (s *Storage) AddPoints(cid int64, user string, points int) error {
	_, err := s.conn.Exec(
		context.Background(),
		`INSERT INTO users (channel_id, username, points) VALUES ($1, lower($2), $3) 
		ON CONFLICT ON CONSTRAINT unique_user_in_channel
			DO UPDATE SET points = users.points + $3, last_activity = current_timestamp`,
		cid, user, points,
	)
	if err != nil {
		return fmt.Errorf("unable to set channel points: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"cid":          cid,
		"user":         user,
		"added_points": points,
	}).Info("adding points")
	return nil
}

func (s *Storage) DeletePoints(channelID int64, user string) error {
	_, err := s.conn.Exec(context.Background(), "DELETE FROM users WHERE channel_id = $1 AND username = lower($2)", channelID, user)
	if err != nil {
		return fmt.Errorf("unable to delete user data: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"cid":  channelID,
		"user": user,
	}).Info("deleting user data")
	return nil
}

func (s *Storage) GetPoints(cid int64, user string) int64 {
	row := s.conn.QueryRow(context.Background(), "SELECT points FROM users WHERE channel_id = $1 AND username = lower($2)", cid, user)
	var points int64
	if err := row.Scan(&points); err != nil {
		logrus.WithError(err).Warn("unable to load user points")
		return 0
	}

	return points
}

func (s *Storage) ListPoints(cid int64, limit int) []StorageResponse {
	channelRow := s.conn.QueryRow(context.Background(), "SELECT channel_name FROM channels WHERE id = $1", cid)
	var channelName string
	if err := channelRow.Scan(&channelName); err != nil {
		logrus.WithError(err).Error("unable to load channel name")
		return []StorageResponse{}
	}

	userRows, err := s.conn.Query(context.Background(), "SELECT username, points FROM users WHERE channel_id = $1 ORDER BY points DESC LIMIT $2", cid, limit)
	if err != nil {
		logrus.WithError(err).Error("unable to load user data")
		return []StorageResponse{}
	}

	response := make([]StorageResponse, 0, limit)
	for userRows.Next() {
		var (
			username string
			points   int64
		)
		userRows.Scan(&username, &points)
		response = append(response, StorageResponse{
			ChannelName: channelName,
			Username:    username,
			Points:      points,
		})
	}

	return response
}

func (s *Storage) AddPPC(cid int64, username string, subscriber bool) error {
	pointUpdtStmt := "points = points + points_per_chat FROM channels "
	if subscriber {
		pointUpdtStmt = "points = points + subscriber_points_per_chat FROM channels "
	}
	cmd, err := s.conn.Exec(
		context.Background(),
		fmt.Sprintf(`UPDATE users SET 
			last_activity = current_timestamp,
			%s
		WHERE channels.id = $1 AND users.channel_id = $1 AND users.username = $2`, pointUpdtStmt),
		cid,
		username,
	)
	if err != nil {
		return fmt.Errorf("unable to delete user data: %w", err)
	}

	if cmd.RowsAffected() <= 0 {
		row := s.conn.QueryRow(context.Background(), "SELECT points_per_chat, subscriber_points_per_chat FROM channels WHERE id = $1", cid)
		var (
			ppc           int64
			subscriberPPC int64
		)
		if err := row.Scan(&ppc, &subscriberPPC); err != nil {
			return fmt.Errorf("unable to load ppc: %w", err)
		}

		initialPoints := ppc
		if subscriber {
			initialPoints = subscriberPPC
		}

		s.conn.Exec(
			context.Background(),
			"INSERT INTO users ( channel_id, username, points ) VALUES ( $1, $2, $3 )",
			cid,
			username,
			initialPoints,
		)
		if err != nil {
			return fmt.Errorf("unable to insert user data: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"cid":  cid,
		"user": username,
	}).Debug("add PPC")
	return nil
}
