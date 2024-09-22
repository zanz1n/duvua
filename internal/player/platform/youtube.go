package platform

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

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
// The youtube search is extremely unstable and may break on any youtube update.
//
// The data is extracted from html youtube responses.
func (y *Youtube) SearchString(s string) (*player.TrackData, error) {
	start := time.Now()
	defer func() {
		slog.Debug(
			"Youtube string search: finished search",
			"took", time.Since(start).Round(time.Microsecond),
		)
	}()

	searchUrl := "https://www.youtube.com/results?search_query=" +
		url.QueryEscape(s)

	req, err := http.NewRequest(http.MethodGet, searchUrl, nil)
	if err != nil {
		return nil, errors.Unexpected("youtube search: request: " + err.Error())
	}

	req.Header.Add("Accept-Language", "en")

	res, err := y.c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Unexpected("youtube search: request: " + err.Error())
	}

	if res.StatusCode != 200 {
		return nil, errors.Unexpected(
			"youtube search: response status: " + res.Status,
		)
	}
	defer res.Body.Close()

	data, err := ytParseSearchBody(res.Body, 1)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errcodes.ErrTrackSearchFailed
	}

	return &data[0], nil
}

// SearchUrl implements Platform.
func (y *Youtube) SearchUrl(url string) ([]player.TrackData, error) {
	if !strings.Contains(url, "&list") && !strings.Contains(url, "?list") {
		video, err := y.c.GetVideo(url)
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
		playlist, err := y.c.GetPlaylist(url)
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
func (y *Youtube) Fetch(id string) (Streamer, error) {
	v, err := y.c.GetVideo("https://www.youtube.com/watch?v=" + id)
	if err != nil {
		return nil, errors.Unexpected(
			"fetch youtube video: " + err.Error(),
		)
	}

	format := filterYtVideos(v.Formats)
	if format == nil {
		return nil, errcodes.ErrTrackSearchFailed
	}

	r, _, err := y.c.GetStream(v, format)
	if err != nil {
		return nil, errors.Unexpected(
			"fetch youtube audio stream: " + err.Error(),
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

func ytParseSearchBody(r io.Reader, limit int) ([]player.TrackData, error) {
	start := time.Now()
	defer func() {
		slog.Debug(
			"Youtube string search: finished parsing data",
			"took", time.Since(start).Round(time.Microsecond),
		)
	}()

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Unexpected(
			"youtube search: response read: " + err.Error(),
		)
	}

	var ok bool

	if bytes.Contains(body, []byte(`var ytInitialData = `)) {
		_, body, ok = bytes.Cut(body, []byte(`var ytInitialData = `))
	} else {
		_, body, ok = bytes.Cut(body, []byte(`window["ytInitialData"] = `))
	}

	if !ok {
		return nil, errors.Unexpected(
			"youtube search: response parse: invalid data",
		)
	}

	dec := json.NewDecoder(bytes.NewReader(body))

	var data ytJsonSearchResult
	if err := dec.Decode(&data); err != nil {
		return nil, errors.Unexpected(
			"youtube search: response parse: " + err.Error(),
		)
	}

	parsed, err := data.Into(limit)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}
