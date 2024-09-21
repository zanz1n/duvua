package anime

import (
	"time"

	"github.com/zanz1n/duvua/internal/utils"
)

type AnimeResourceResponse struct {
	Data Anime `json:"data"`
}

type AnimeCollectionResponse struct {
	Data []Anime `json:"data"`
	Meta struct {
		Count int32 `json:"count"`
	}
	Links struct {
		First string `json:"first" validate:"required"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
		Last  string `json:"last" validate:"required"`
	}
}

type Anime struct {
	ID    utils.StringInt `json:"id" validate:"required"`
	Type  string          `json:"type" validate:"required"`
	Links struct {
		Self string `json:"self" validate:"required"`
	} `json:"links"`
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
	Slug      string    `json:"string"`
	// The synopsis of the anime in english
	Synopsis string `json:"synopsis" validate:"required"`
	// Titles in different languages
	Titles            AnimeTitles `json:"titles"`
	CanonicalTitle    string      `json:"canonicalTitle"`
	AbbreviatedTitles []string    `json:"abbreviatedTitles"`

	AverageRating     string            `json:"averageRating"`
	RatingFrequencies map[string]string `json:"ratingFrequencies"`
	UserCount         int32             `json:"userCount"`
	FavoritesCount    int32             `json:"favoritesCount"`
	PopularityRank    int32             `json:"popularityRank"`
	RatingRank        int32             `json:"ratingRank"`

	StartDate Date `json:"startDate"`
	EndDate   Date `json:"endDate"`

	AgeRating      AnimeAgeRating `json:"ageRating" validate:"required"`
	AgeRatingGuide string         `json:"ageRatingGuide"`

	Subtype       AnimeSubtype `json:"subtype" validate:"required"`
	Status        AnimeStatus  `json:"status" validate:"required"`
	TBA           string       `json:"tba"`
	EpisodeCount  int32        `json:"episodeCount"`
	EpisodeLength int32        `json:"episodeLength"`
	NSFW          bool         `json:"nsfw"`

	PosterImage    AnimeImage `json:"posterImage"`
	CoverImage     AnimeImage `json:"coverImage"`
	YoutubeVideoId string     `json:"youtubeVideoId"`
}

type AnimeTitles struct {
	English         string `json:"en"`
	EnglishJapanese string `json:"en_jp"`
	JapanJapanese   string `json:"ja_jp"`
}

type AnimeImage struct {
	Tiny     string `json:"tiny"`
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Original string `json:"original"`

	Meta struct {
		Dimensions AnimeImageDimensions `json:"dimensions"`
	} `json:"meta"`
}

type AnimeImageDimensions struct {
	Tiny   AnimeImageDimension `json:"tiny"`
	Small  AnimeImageDimension `json:"small"`
	Medium AnimeImageDimension `json:"medium"`
	Large  AnimeImageDimension `json:"large"`
}

type AnimeImageDimension struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}
