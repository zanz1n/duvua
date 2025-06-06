package platform

import (
	"io"
	"net/url"
	"strings"

	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/player/encoder"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/pkg/pb/player"
)

type Fetcher struct {
	yt Platform
	sp Platform
}

func NewFetcher(ytf *Youtube, spf *Spotify) *Fetcher {
	if ytf == nil {
		ytf = NewYoutube(nil, 1)
	}
	return &Fetcher{yt: ytf, sp: spf}
}

func (f *Fetcher) Search(query string) ([]*player.TrackData, error) {
	if strings.HasPrefix(query, "https://") {
		u, err := url.Parse(query)
		if err != nil {
			return nil, errcodes.ErrTrackSearchInvalidUrl
		}

		switch {
		case strings.Contains(u.Host, "youtu"):
			return f.yt.SearchUrl(query)

		case strings.Contains(u.Host, "spotify"):
			return f.sp.SearchUrl(query)

		// case strings.Contains(u.Host, "soundcloud"):
		default:
			return nil, errcodes.ErrTrackSearchUnsuported
		}
	}

	track, err := f.yt.SearchString(query)
	if err != nil {
		return nil, err
	}

	return []*player.TrackData{track}, nil
}

func (f *Fetcher) Fetch(query string) (Streamer, error) {
	platform, id, ok := strings.Cut(query, ":")
	if !ok {
		return nil, errors.New("invalid music format")
	}

	switch platform {
	case "youtube":
		return f.yt.Fetch(id)

	// case "spotify":
	// case "soundcloud":
	default:
		return nil, errors.New("invalid format")
	}
}

var _ Streamer = &readerStreamer{}

type readerStreamer struct {
	*encoder.Session
}

func newReaderStreamer(r io.ReadCloser) (*readerStreamer, error) {
	return &readerStreamer{encoder.NewSession(r, nil)}, nil
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
