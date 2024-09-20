package platform

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kkdai/youtube/v2"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/pkg/player"
)

var _ Platform = &Youtube{}

type Youtube struct {
	c        *youtube.Client
	hc       *http.Client
	validate *validator.Validate
}

// if client == nil, it will be defaulted to http.DefaultClient
func NewYoutube(client *http.Client, maxDlRoutines int) *Youtube {
	if client == nil {
		client = http.DefaultClient
	}

	return &Youtube{
		c: &youtube.Client{
			HTTPClient:  client,
			MaxRoutines: maxDlRoutines,
		},
		hc:       client,
		validate: validator.New(),
	}
}

// SearchString implements Platform.
func (f *Youtube) SearchString(s string) (*player.TrackData, error) {
	panic("unimplemented")
}

// SearchUrl implements Platform.
func (f *Youtube) SearchUrl(url string) ([]player.TrackData, error) {
	if !strings.Contains(url, "&list") && !strings.Contains(url, "?list") {
		video, err := f.c.GetVideo(url)
		if err != nil {
			return nil, errcodes.ErrTrackSearchFailed
		}

		thumbnailUrl := defaultThumbUrl
		if thumbnail := filterYtThumbnails(video.Thumbnails); thumbnail != nil {
			thumbnailUrl = thumbnail.URL
		}

		return []player.TrackData{{
			Name:      video.Title,
			URL:       "https://youtu.be/" + video.ID,
			PlayQuery: "youtube:" + video.ID,
			Thumbnail: thumbnailUrl,
			Duration:  video.Duration,
		}}, nil
	} else {
		playlist, err := f.c.GetPlaylist(url)
		if err != nil {
			return nil, errcodes.ErrTrackSearchFailed
		}

		tracks := make([]player.TrackData, len(playlist.Videos))
		for i, video := range playlist.Videos {
			thumbnailUrl := defaultThumbUrl
			if thumbnail := filterYtThumbnails(video.Thumbnails); thumbnail != nil {
				thumbnailUrl = thumbnail.URL
			}

			tracks[i] = player.TrackData{
				Name:      video.Title,
				URL:       "https://youtu.be/" + video.ID,
				PlayQuery: "youtube:" + video.ID,
				Thumbnail: thumbnailUrl,
				Duration:  video.Duration,
			}
		}

		if len(tracks) == 0 {
			return nil, errcodes.ErrTrackSearchFailed
		}

		return tracks, nil
	}
}

// Fetch implements Platform.
func (f *Youtube) Fetch(id string) (Streamer, error) {
	v, err := f.c.GetVideo("https://www.youtube.com/watch?v=" + id)
	if err != nil {
		return nil, errors.Unexpected(
			"failed to fetch youtube video: " + err.Error(),
		)
	}

	format := filterYtVideos(v.Formats)
	if format == nil {
		return nil, errcodes.ErrTrackSearchFailed
	}

	r, _, err := f.c.GetStream(v, format)
	if err != nil {
		return nil, errors.Unexpected(
			"failed to fetch youtube audio stream: " + err.Error(),
		)
	}

	return newReaderStreamer(r)
}

func filterYtVideos(formats []youtube.Format) *youtube.Format {
	find := -1
	for i, f := range formats {
		if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" &&
			strings.Contains(f.MimeType, "audio") &&
			strings.Contains(f.MimeType, "opus") {
			find = i
		}
	}

	if find == -1 {
		for i, f := range formats {
			if f.AudioQuality == "AUDIO_QUALITY_MEDIUM" &&
				strings.Contains(f.MimeType, "audio") &&
				strings.Contains(f.MimeType, "mp4") {
				find = i
			}
		}
	}
	if find == -1 {
		return nil
	}

	return &formats[find]
}

func filterYtThumbnails(thumbnails []youtube.Thumbnail) *youtube.Thumbnail {
	if len(thumbnails) > 0 {
		return &thumbnails[0]
	}
	return nil
}
