package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type TrackData struct {
	TrackName  string `json:"track_name"`
	ArtistName string `json:"artist_name"`
}

func GetTrackPostHandler(conn *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID, err := getChannelIdByToken(conn, c.Param("token"))
		if err != nil {
			logrus.WithError(err).Error("unable to load channel ID")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		payload := &TrackData{}
		if err := c.BindJSON(payload); err != nil {
			logrus.WithError(err).Error("unable to parse track data to update")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := setCurrentTrack(conn, channelID, payload.TrackName, payload.ArtistName); err != nil {
			logrus.WithError(err).Error("")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Status(http.StatusOK)
	}
}

func GetTrackGetHanmdler(conn *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelID, err := getChannelIdByToken(conn, c.Param("token"))
		if err != nil {
			logrus.WithError(err).Error("unable to load channel ID")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		trackName, artistName, err := getCurrentTrack(conn, channelID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			} else {
				logrus.WithError(err).Error("unable to load track data")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}

		c.JSON(http.StatusOK, TrackData{
			TrackName:  trackName,
			ArtistName: artistName,
		})
	}
}

func getChannelIdByToken(conn *pgxpool.Pool, token string) (int64, error) {
	channelrow := conn.QueryRow(context.Background(), "SELECT id FROM channels WHERE token = $1", token)
	var channelID int64
	if err := channelrow.Scan(&channelID); err != nil {
		return 0, fmt.Errorf("unable to load channel ID: %w", err)
	}

	return channelID, nil
}

func setCurrentTrack(conn *pgxpool.Pool, channelID int64, trackName, artistName string) error {
	_, err := conn.Exec(
		context.Background(),
		`INSERT INTO current_track ( channel_id, track_name, artist_name ) VALUES ( $1, $2, $3 )
		 ON CONFLICT ON CONSTRAINT unique_per_channel
			DO UPDATE SET track_name = $2, artist_name = $3`,
		channelID,
		trackName,
		artistName,
	)
	if err != nil {
		return fmt.Errorf("unable to save track: %w", err)
	}

	return nil
}

func getCurrentTrack(conn *pgxpool.Pool, channelID int64) (string, string, error) {
	var (
		trackName  string
		artistName string
	)
	row := conn.QueryRow(context.Background(), "SELECT track_name, artist_name FROM current_track WHERE channel_id = $1", channelID)
	if err := row.Scan(&trackName, &artistName); err != nil {
		return "", "", fmt.Errorf("unable to load track: %w", err)
	}

	return trackName, artistName, nil
}
