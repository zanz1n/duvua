package platform

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"sync/atomic"
	"time"

	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/pkg/pb/player"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var _ Platform = &Spotify{}

type Spotify struct {
	c  atomic.Pointer[spotify.Client]
	yt *Youtube

	cfg clientcredentials.Config
}

func NewSpotify(clientId, clientSecret string, yt *Youtube) (*Spotify, error) {
	s := &Spotify{
		yt: yt,
		cfg: clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     spotifyauth.TokenURL,
		},
	}

	if err := s.rollAuth(); err != nil {
		return nil, errors.Unexpected("spotify auth: " + err.Error())
	}

	return s, nil
}

// SearchString implements Platform.
func (s *Spotify) SearchString(query string) (*player.TrackData, error) {
	res, err := authRetry(s, func(c *spotify.Client) (*spotify.SearchResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		return c.Search(
			ctx,
			query,
			spotify.SearchTypeTrack,
			spotify.Limit(2),
		)
	})

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
func (s *Spotify) SearchUrl(uri string) ([]*player.TrackData, error) {

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

	switch pathS[1] {
	case "track":
		track, err := authRetry(s, func(c *spotify.Client) (*spotify.FullTrack, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			return c.GetTrack(ctx, spotify.ID(id))
		})

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

		return []*player.TrackData{data}, nil

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

func (s *Spotify) rollAuth() error {
	start := time.Now()

	token, err := s.cfg.Token(context.Background())
	if err != nil {
		slog.Error(
			"Spotify: Failed to reroll auth token",
			"took", time.Since(start).Round(time.Millisecond),
			"error", err,
		)
		return err
	}

	hc := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(hc)

	s.c.Store(client)
	slog.Info(
		"Spotify: Rerolled auth token",
		"took", time.Since(start).Round(time.Millisecond),
	)

	return nil
}

func authRetry[T any](s *Spotify, f func(c *spotify.Client) (*T, error)) (*T, error) {
	if s.c.Load() == nil {
		if err := s.rollAuth(); err != nil {
			return nil, err
		}
	}

	t, err := f(s.c.Load())

	if err != nil && isExpiredTokenErr(err) {
		if err2 := s.rollAuth(); err2 != nil {
			return nil, err2
		}

		t, err = f(s.c.Load())
	}

	return t, err
}

func (s *Spotify) ytConvert(track *spotify.FullTrack) (*player.TrackData, error) {
	ytSearchQuery := ""
	if len(track.Artists) > 0 {
		ytSearchQuery += track.Artists[0].Name + " "
	}
	ytSearchQuery += track.Name

	return s.yt.SearchString(ytSearchQuery)
}

func isExpiredTokenErr(err error) bool {
	return strings.Contains(err.Error(), "token expired")
}
