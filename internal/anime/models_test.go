package anime_test

import (
	"embed"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/zanz1n/duvua/internal/anime"
)

//go:embed *.json
var embedfs embed.FS

func TestAnimeDecodeEncode(t *testing.T) {
	validate := validator.New()

	file, err := embedfs.ReadFile("test_data.json")
	assert.Nil(t, err, "Failed to open test_data.json file")

	a := anime.Anime{}
	err = json.Unmarshal(file, &a)
	assert.Nil(t, err, "Failed to unmarshal test_json data to anime.Anime struct")

	err = validate.Struct(&a)
	assert.Nil(t, err, "Failed to validate unmarshaled anime data")

	_, err = json.Marshal(&a)
	assert.Nil(t, err, "Failed to marshal anime.Anime struct")
}
