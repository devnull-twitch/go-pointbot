package pointbot

import (
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	Action          int
	StorageResponse struct {
		ChannelName string
		Username    string
		Points      int
	}
	StorageRequest struct {
		Action      Action
		ChannelName string
		Token       string
		Username    string
		Points      int
		ReplyChan   chan StorageResponse
	}
	channelBlock struct {
		Name   string
		Points map[string]int
	}
	Storage struct {
		channelStorage  map[string]*channelBlock
		channelTokenMap map[string]string
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
)

func NewStorage() chan<- StorageRequest {
	requestChan := make(chan StorageRequest)
	go func(reqs <-chan StorageRequest) {
		s := &Storage{
			channelStorage:  make(map[string]*channelBlock),
			channelTokenMap: make(map[string]string),
		}

		for {
			req := <-reqs

			if req.Token == "" && req.ChannelName != "" {
				req.Token = s.channelTokenMap[req.ChannelName]
			}

			switch req.Action {
			case ActionJoin:
				s.AddChannel(req.Token, req.ChannelName)

			case ActionAddPoints:
				if err := s.AddPoints(req.Token, req.Username, req.Points); err != nil {
					logrus.WithError(err).Error("unable to add points")
				}

			case ActionDelPoints:
				if err := s.DeletePoints(req.Token, req.Username); err != nil {
					logrus.WithError(err).Error("unable to delete points")
				}

			case ActionSubPoints:
				if err := s.AddPoints(req.Token, req.Username, -req.Points); err != nil {
					logrus.WithError(err).Error("unable to subtract points")
				}

			case ActionGetPoints:
				points := s.GetPoints(req.Token, req.Username)
				select {
				case req.ReplyChan <- StorageResponse{Username: req.Username, ChannelName: req.ChannelName, Points: points}:
				case <-time.After(time.Second):
					logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
				}

			case ActionTop:
				resps := s.ListPoints(req.Token, 1)
				if len(resps) <= 0 {
					req.ReplyChan <- StorageResponse{}
					return
				}

				select {
				case req.ReplyChan <- resps[0]:
				case <-time.After(time.Second):
					logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
				}

			case ActionList:
				resps := s.ListPoints(req.Token, 10)
				if len(resps) <= 0 {
					req.ReplyChan <- StorageResponse{}
					return
				}

				for _, r := range resps {
					select {
					case req.ReplyChan <- r:
					case <-time.After(time.Second):
						logrus.WithField("channel", req.ChannelName).Warn("bot didnt take response")
						return
					}
				}

				close(req.ReplyChan)
			}
		}
	}(requestChan)

	return requestChan
}

func (s *Storage) AddChannel(token, channel string) {
	s.channelStorage[token] = &channelBlock{
		Name:   channel,
		Points: make(map[string]int),
	}
	s.channelTokenMap[channel] = token
}

func (s *Storage) AddPoints(token, user string, points int) error {
	channel, exists := s.channelStorage[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	channel.Points[user] += points
	logrus.WithFields(logrus.Fields{
		"channel":      channel.Name,
		"user":         user,
		"total_points": channel.Points[user],
		"added_points": points,
	}).Info("adding points")
	return nil
}

func (s *Storage) DeletePoints(token, user string) error {
	channel, exists := s.channelStorage[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	delete(channel.Points, user)
	logrus.WithFields(logrus.Fields{
		"channel": channel.Name,
		"user":    user,
	}).Info("deleting user data")
	return nil
}

func (s *Storage) GetPoints(token, user string) int {
	channel, exists := s.channelStorage[token]
	if !exists {
		logrus.Error("bot is in channel but also not")
		return 0
	}

	return channel.Points[user]
}

func (s *Storage) ListPoints(token string, limit int) []StorageResponse {
	channel, exists := s.channelStorage[token]
	if !exists {
		logrus.Error("bot is in channel but also not")
		return nil
	}

	response := make([]StorageResponse, 0, len(channel.Points))
	for user, points := range channel.Points {
		response = append(response, StorageResponse{
			ChannelName: channel.Name,
			Username:    user,
			Points:      points,
		})
	}

	sort.Slice(response, func(i, j int) bool {
		return response[i].Points > response[j].Points
	})

	if len(response) > limit {
		return response[0:limit]
	}

	return response
}
