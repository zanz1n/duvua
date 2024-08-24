package anime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zanz1n/duvua-bot/internal/errors"
)

func NewAnimeApi(client *http.Client) *AnimeApi {
	if client == nil {
		client = http.DefaultClient
	}

	return &AnimeApi{
		Client:   client,
		Timeout:  2 * time.Second,
		validate: validator.New(),
	}
}

type AnimeApi struct {
	Client   *http.Client
	Timeout  time.Duration
	validate *validator.Validate
}

func (a *AnimeApi) FetchByName(name string, limit, offset int) (*AnimeCollectionResponse, error) {
	url := "https://kitsu.io/api/edge/anime"

	ctx, cancel := context.WithTimeout(context.Background(), a.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Unexpected("anime fetch: " + err.Error())
	}

	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Content-Type", "application/vnd.api+json")

	values := req.URL.Query()
	values.Add("filter[text]", name)
	values.Add("page[limit]", strconv.Itoa(limit))
	values.Add("page[offset]", strconv.Itoa(offset))
	values.Add("sort", "popularityRank")
	req.URL.RawQuery = values.Encode()

	res, err := a.Client.Do(req)
	if err != nil {
		return nil, errors.Unexpected("anime fetch: " + err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, errors.New("anime não encontrado")
		default:
			return nil, errors.Unexpectedf(
				"anime fetch: unexpected http status %s",
				res.Status,
			)
		}
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Unexpectedf(
			"anime fetch: failed to read response body: %s",
			err,
		)
	}

	anime := AnimeCollectionResponse{}
	if err = json.Unmarshal(b, &anime); err != nil {
		return nil, errors.Unexpected("anime parse: " + err.Error())
	}
	if err = a.validate.Struct(&anime); err != nil {
		return nil, errors.Unexpected("anime validate: " + err.Error())
	}

	return &anime, nil
}

func (a *AnimeApi) GetByName(name string) (*Anime, error) {
	res, err := a.FetchByName(name, 1, 0)
	if err != nil {
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, errors.New("anime não encontrado")
	}

	return &res.Data[0], nil
}

func (a *AnimeApi) GetById(id int64) (*Anime, error) {
	url := fmt.Sprintf("https://kitsu.io/api/edge/anime/%d", id)

	ctx, cancel := context.WithTimeout(context.Background(), a.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Unexpected("anime fetch: " + err.Error())
	}

	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Content-Type", "application/vnd.api+json")

	res, err := a.Client.Do(req)
	if err != nil {
		return nil, errors.Unexpected("anime fetch: " + err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, errors.New("anime não encontrado")
		default:
			return nil, errors.Unexpectedf(
				"anime fetch: unexpected http status %s",
				res.Status,
			)
		}
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Unexpectedf(
			"anime fetch: failed to read response body: %s",
			err,
		)
	}

	anime := AnimeResourceResponse{}
	if err = json.Unmarshal(b, &anime); err != nil {
		return nil, errors.Unexpected("anime parse: " + err.Error())
	}
	if err = a.validate.Struct(&anime); err != nil {
		return nil, errors.Unexpected("anime validate: " + err.Error())
	}

	return &anime.Data, nil
}
