package platform

import (
	"io"

	"github.com/zanz1n/duvua/pkg/player"
)

type TrackSpeed int8

const (
	TrackSpeed_0_25X TrackSpeed = iota - 3
	TrackSpeed_0_5X
	TrackSpeed_0_75X
	// Default speed
	TrackSpeed_1X
	TrackSpeed_1_25X
	TrackSpeed_1_5X
	TrackSpeed_1_75X
	TrackSpeed_2X
)

type Streamer interface {
	ReadOpus() ([]byte, error)
	SetSpeed(speed TrackSpeed) error
	SetVolume(volume uint8) error

	io.Closer
}

type Platform interface {
	SearchString(s string) (*player.TrackData, error)
	SearchUrl(url string) (*player.TrackData, error)
	Fetch(url string) (Streamer, error)
}
