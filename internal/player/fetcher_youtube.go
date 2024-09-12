package player

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kkdai/youtube/v2"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/pkg/player"
)

var _ PlatformFetcher = &YoutubeFetcher{}

type YoutubeFetcher struct {
	c        *youtube.Client
	hc       *http.Client
	validate *validator.Validate
}

// if client == nil, it will be defaulted to http.DefaultClient
func NewYoutubeFetcher(client *http.Client, maxDlRoutines int) *YoutubeFetcher {
	if client == nil {
		client = http.DefaultClient
	}

	return &YoutubeFetcher{
		c: &youtube.Client{
			HTTPClient:  client,
			MaxRoutines: maxDlRoutines,
		},
		hc:       client,
		validate: validator.New(),
	}
}

func (f *YoutubeFetcher) SearchString(s string) (*player.TrackData, error) {
	panic("unimplemented")
}

func (f *YoutubeFetcher) SearchUrl(url string) (*player.TrackData, error) {
	v, err := f.c.GetVideo(url)
	if err != nil {
		return nil, ErrTrackSearchFailed
	}
	yturl := "https://www.youtube.com/watch?v=" + v.ID

	thumbnailUrl := "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRd2NAjCcjjk7ac57mKCQvgWVTmP0ysxnzQnQ&s"
	if thumbnail := filterYtThumbnails(v.Thumbnails); thumbnail != nil {
		thumbnailUrl = thumbnail.URL
	}

	return &player.TrackData{
		Name:      v.Title,
		URL:       yturl,
		PlayQuery: "youtube:" + yturl,
		Thumbnail: thumbnailUrl,
		Duration:  v.Duration,
	}, nil
}

func (f *YoutubeFetcher) Fetch(url string) (Streamer, error) {
	v, err := f.c.GetVideo(url)
	if err != nil {
		return nil, errors.Unexpected(
			"failed to fetch youtube video: " + err.Error(),
		)
	}

	format := filterYtVideos(v.Formats)
	if format == nil {
		return nil, ErrTrackSearchFailed
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
