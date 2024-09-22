package platform

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/player"
)

type ytVideo struct {
	VideoId string `json:"videoId"`

	// All the available sizes of the youtube thumbnail
	Thumbnail struct {
		Thumbnails []ytThumbnail `json:"thumbnails"`
	} `json:"thumbnail"`

	// Title of the video splited in rich text
	Title ytRichText `json:"title"`

	// The text containing the text formated duration of the videos
	LengthText struct {
		SimpleText string `json:"simpleText"`
	} `json:"lengthText"`

	// The text containing the text formated number of views
	// ViewCountText struct {
	// 	SimpleText string `json:"simpleText"`
	// } `json:"viewCountText"`

	// The name of the channel that published the video
	// OwnerText ytRichText `json:"ownerText"`

	// // The inner text contains the shorted youtube description
	// DetailedMetadataSnippets [1]struct {
	// 	SnippetText ytRichText `json:"snippetText"`
	// } `json:"detailedMetadataSnippets"`
}

func (v *ytVideo) Into() (player.TrackData, bool) {
	thumbnail := defaultThumbUrl
	if len(v.Thumbnail.Thumbnails) > 0 {
		thumbnail = v.Thumbnail.Thumbnails[0].URL
	}

	ok := true

	title := v.Title.String()
	if title == "" {
		ok = false
	}

	if v.VideoId == "" {
		ok = false
	}

	var duration time.Duration
	dsplit := strings.Split(v.LengthText.SimpleText, ":")
	slices.Reverse(dsplit)

	if len(dsplit) >= 3 {
		hours, _ := strconv.ParseInt(dsplit[2], 10, 0)
		duration += (time.Duration(hours) * time.Hour)
	}
	if len(dsplit) >= 2 {
		minutes, _ := strconv.ParseInt(dsplit[1], 10, 0)
		duration += (time.Duration(minutes) * time.Minute)
	}
	if len(dsplit) >= 1 {
		seconds, _ := strconv.ParseInt(dsplit[0], 10, 0)
		duration += (time.Duration(seconds) * time.Second)
	}

	if duration == 0 {
		ok = false
	}

	return player.TrackData{
		Name:      title,
		URL:       "https://youtu.be/" + v.VideoId,
		PlayQuery: "youtube:" + "https://www.youtube.com/watch?v=" + v.VideoId,
		Thumbnail: thumbnail,
		Duration:  duration,
	}, ok
}

type ytThumbnail struct {
	URL    string `json:"url"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
}

func (t ytThumbnail) Into() youtube.Thumbnail {
	return youtube.Thumbnail{
		URL:    t.URL,
		Width:  t.Width,
		Height: t.Height,
	}
}

// The rich text represents a text with bold, italic, etc...
type ytRichText struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

func (t ytRichText) String() (s string) {
	for _, unit := range t.Runs {
		s += unit.Text
	}
	return
}

type ytSearchResultContent struct {
	VideoRenderer *ytVideo `json:"videoRenderer"`
}

type ytJsonSearchResult struct {
	ResultsCount utils.StringInt `json:"estimatedResults"`
	Contents     struct {
		TwoColumnSearchResultsRenderer struct {
			PrimaryContents struct {
				SectionListRenderer struct {
					Contents [1]struct {
						ItemSectionRenderer struct {
							Contents []rawJsonBytes `json:"contents"`
						} `json:"itemSectionRenderer"`
					} `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
	} `json:"contents"`
}

func (r *ytJsonSearchResult) Into(limit int) ([]player.TrackData, error) {
	videosRaw := r.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.
		SectionListRenderer.Contents[0].ItemSectionRenderer.Contents

	videos := []player.TrackData{}

	for i, content := range videosRaw {
		if len(videos) > limit {
			break
		}

		var data ytSearchResultContent
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, errors.Unexpectedf(
				"youtube search: parse contents at [%d]: "+
					err.Error(), i,
			)
		}

		if data.VideoRenderer != nil {
			data, ok := data.VideoRenderer.Into()
			if ok {
				videos = append(videos, data)
			}
		}
	}

	return videos, nil
}

type rawJsonBytes []byte

// UnmarshalJSON implements json.Unmarshaler.
func (rb *rawJsonBytes) UnmarshalJSON(b []byte) error {
	*rb = b
	return nil
}
