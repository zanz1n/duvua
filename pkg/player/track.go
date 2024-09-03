package player

import (
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	RequestedAt time.Time `json:"requested_at" validate:"required"`
	UserId      uint64    `json:"user_id" validate:"required"`
	ChannelId   uint64    `json:"channel_id" validate:"required"`

	State *TrackState `json:"state"`
	Data  *TrackData  `json:"data" validate:"required"`
}

func NewTrack(userId, channelId uint64, data *TrackData) Track {
	return Track{
		ID:          uuid.New(),
		RequestedAt: time.Now(),
		State:       nil,
		UserId:      userId,
		ChannelId:   channelId,
		Data:        data,
	}
}

type TrackState struct {
	Progress     time.Duration `json:"progress"`
	PlayingStart time.Time     `json:"play_start" validate:"required"`
	Looping      bool          `json:"looping"`
}

type TrackData struct {
	Name      string        `json:"name" validate:"required"`
	URL       string        `json:"url" validate:"required"`
	PlayQuery string        `json:"play_query" validate:"required"`
	Thumbnail string        `json:"thumbnail" validate:"required"`
	Duration  time.Duration `json:"duration" validate:"required"`
}

type AddTrackData struct {
	UserId    string `json:"user_id" validate:"required,number"`
	ChannelId string `json:"channel_id" validate:"required,number"`

	Data *TrackData `json:"data" validate:"required"`
}
