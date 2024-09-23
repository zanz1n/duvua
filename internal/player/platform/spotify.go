package platform

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/pkg/player"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var _ Platform = &Spotify{}

type Spotify struct {
	c  *spotify.Client
	yt *Youtube
}

func NewSpotify(clientId, clientSecret string, yt *Youtube) (*Spotify, error) {
	config := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(context.Background())
	if err != nil {
		return nil, err
	}

	hc := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(hc)

	return &Spotify{
		c:  client,
		yt: yt,
	}, nil
}

// SearchString implements Platform.
func (s *Spotify) SearchString(query string) (*player.TrackData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := s.c.Search(
		ctx,
		query,
		spotify.SearchTypeTrack,
		spotify.Limit(2),
	)
	if err != nil {
		return nil, errors.Unexpected(
			"spotify search string: " + err.Error(),
		)
	}

	if res.Tracks == nil || len(res.Tracks.Tracks) == 0 {
		return nil, errcodes.ErrTrackSearchFailed
	}

	track := res.Tracks.Tracks[0]

	return s.ytConvert(&track)
}

// SearchUrl implements Platform.
func (s *Spotify) SearchUrl(uri string) ([]player.TrackData, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return nil, errcodes.ErrTrackSearchFailed
	}

	pathS := strings.Split(u.Path, "/")
	if 3 > len(pathS) {
		return nil, errcodes.ErrTrackSearchFailed
	}

	id := pathS[len(pathS)-1]
	pathS = pathS[:len(pathS)-1]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	switch pathS[1] {
	case "track":
		track, err := s.c.GetTrack(ctx, spotify.ID(id))
		if err != nil {
			return nil, errors.Unexpected(
				"spotify search track: " + err.Error(),
			)
		}

		start := time.Now()
		data, err := s.ytConvert(track)
		if err != nil {
			slog.Warn(
				"Spotify: Failed to search spotify track on youtube",
				"took", time.Since(start),
				"error", err,
			)
			return nil, err
		}

		return []player.TrackData{*data}, nil

	case "playlist", "album":
		return nil, errcodes.ErrSpotifyPlaylistsNotSupported

	default:
		return nil, errcodes.ErrTrackSearchFailed
	}
}

// Fetch implements Platform.
func (s *Spotify) Fetch(query string) (Streamer, error) {
	panic("must not be used")
}

func (s *Spotify) ytConvert(track *spotify.FullTrack) (*player.TrackData, error) {
	ytSearchQuery := ""
	if len(track.Artists) > 0 {
		ytSearchQuery += track.Artists[0].Name + " "
	}
	ytSearchQuery += track.Name

	return s.yt.SearchString(ytSearchQuery)
}
