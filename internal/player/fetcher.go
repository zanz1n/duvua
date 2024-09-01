package player

import (
	"io"
	"net/url"
	"strings"

	"github.com/jogramming/dca"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/pkg/player"
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

type PlatformFetcher interface {
	SearchString(s string) (*player.TrackData, error)
	SearchUrl(url string) (*player.TrackData, error)
	Fetch(url string) (Streamer, error)
}

type TrackFetcher struct {
	yt PlatformFetcher
}

func NewTrackFetcher(ytf *YoutubeFetcher) *TrackFetcher {
	if ytf == nil {
		ytf = NewYoutubeFetcher(nil, 1)
	}
	return &TrackFetcher{yt: ytf}
}

func (f *TrackFetcher) Search(query string) (*player.TrackData, error) {
	if strings.HasPrefix(query, "https://") {
		u, err := url.Parse(query)
		if err != nil {
			return nil, errors.New("invalid url: " + err.Error())
		}

		switch {
		case strings.Contains(u.Host, "youtu"):
			return f.yt.SearchUrl(query)

		// case strings.Contains(u.Host, "soundcloud"):
		// case strings.Contains(u.Host, "spotify"):
		default:
			return nil, errors.Newf("invalid url host `%s`", u.Host)
		}
	}

	return f.yt.SearchString(query)
}

func (f *TrackFetcher) Fetch(query string) (Streamer, error) {
	platform, url, ok := strings.Cut(query, ":")
	if !ok {
		return nil, errors.New("invalid music format")
	}

	switch platform {
	case "youtube":
		return f.yt.Fetch(url)

	// case "spotify":
	// case "soundcloud":
	default:
		return nil, errors.New("invalid format")
	}
}

var _ Streamer = &readerStreamer{}

type readerStreamer struct {
	rc   io.ReadCloser
	sess *dca.EncodeSession
}

func newReaderStreamer(r io.ReadCloser) (*readerStreamer, error) {
	opts := &dca.EncodeOptions{
		Volume:           256,
		Channels:         2,
		FrameRate:        48000,
		FrameDuration:    20,
		Bitrate:          64,
		Application:      dca.AudioApplicationAudio,
		CompressionLevel: 10,
		PacketLoss:       1,
		BufferedFrames:   100, // At 20ms frames that's 2s
		VBR:              true,
		StartTime:        0,
		Threads:          1,
	}

	ess, err := dca.EncodeMem(r, opts)
	if err != nil {
		return nil, err
	}

	return &readerStreamer{rc: r, sess: ess}, nil
}

// Close implements Streamer.
func (s *readerStreamer) Close() error {
	s.sess.Cleanup()
	return s.rc.Close()
}

// ReadOpus implements Streamer.
func (s *readerStreamer) ReadOpus() ([]byte, error) {
	return s.sess.OpusFrame()
}

// SetSpeed implements Streamer.
func (s *readerStreamer) SetSpeed(speed TrackSpeed) error {
	// TODO: implement speed
	return nil
}

// SetVolume implements Streamer.
func (s *readerStreamer) SetVolume(volume uint8) error {
	// TODO: implement volume
	return nil
}
