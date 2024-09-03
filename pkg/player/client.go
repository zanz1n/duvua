package player

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zanz1n/duvua-bot/internal/errors"
)

type dataBody struct {
	Message string `json:"message"`
	Changed bool   `json:"changed"`
	Data    any    `json:"data"`
}

type errBody struct {
	Error     string `json:"error" validate:"required"`
	ErrorCode int    `json:"error_code"`
}

type HttpClient struct {
	c        *http.Client
	validate *validator.Validate
	timeout  time.Duration
	passwd   string
	baseUrl  string
}

func NewHttpClient(client *http.Client, baseUrl string, password string) *HttpClient {
	if client == nil {
		client = http.DefaultClient
	}
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl + "/"
	}

	return &HttpClient{
		c:        client,
		validate: validator.New(),
		timeout:  2 * time.Second,
		passwd:   password,
		baseUrl:  baseUrl,
	}
}

func (h *HttpClient) FetchTrack(query string) (*TrackData, error) {
	q := url.Values{}
	q.Add("query", query)

	var resData TrackData

	_, err := h.request(http.MethodGet, "track/search", q, nil, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) AddTrack(guildId string, data AddTrackData) (*Track, error) {
	var resData Track

	url := fmt.Sprintf("guild/%s/track", guildId)
	_, err := h.request(http.MethodPost, url, nil, data, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) GetPlayingTrack(guildId string) (*Track, error) {
	var resData Track

	url := fmt.Sprintf("guild/%s/track", guildId)
	_, err := h.request(http.MethodGet, url, nil, nil, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) Skip(guildId string) (*Track, error) {
	var resData Track

	url := fmt.Sprintf("guild/%s/skip", guildId)
	_, err := h.request(http.MethodPut, url, nil, nil, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) Stop(guildId string) error {
	url := fmt.Sprintf("guild/%s", guildId)
	_, err := h.request(http.MethodDelete, url, nil, nil, nil)
	return err
}

func (h *HttpClient) Pause(guildId string) (bool, error) {
	url := fmt.Sprintf("guild/%s/pause", guildId)
	return h.request(http.MethodPut, url, nil, nil, nil)
}

func (h *HttpClient) Unpause(guildId string) (bool, error) {
	url := fmt.Sprintf("guild/%s/unpause", guildId)
	return h.request(http.MethodPut, url, nil, nil, nil)
}

func (h *HttpClient) EnableLoop(guildId string, enable bool) (bool, error) {
	q := url.Values{}
	q.Add("enable", strconv.FormatBool(enable))

	url := fmt.Sprintf("guild/%s/loop", guildId)
	return h.request(http.MethodPut, url, q, nil, nil)
}

func (h *HttpClient) SetVolume(guildId string, volume uint8) (bool, error) {
	q := url.Values{}
	q.Add("volume", strconv.Itoa(int(volume)))

	url := fmt.Sprintf("guild/%s/volume", guildId)
	return h.request(http.MethodPut, url, q, nil, nil)
}

func (h *HttpClient) GetTrackById(guildId string, id uuid.UUID) (*Track, error) {
	var resData Track

	url := fmt.Sprintf("guild/%s/track/%s", guildId, id.String())
	_, err := h.request(http.MethodGet, url, nil, nil, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) GetTracks(guildId string) ([]Track, error) {
	var resData []Track

	url := fmt.Sprintf("guild/%s/tracks", guildId)
	_, err := h.request(http.MethodGet, url, nil, nil, &resData)
	if err != nil {
		return nil, err
	}

	return resData, nil
}

func (h *HttpClient) RemoveTrack(guildId string, id uuid.UUID) (*Track, error) {
	var resData Track

	url := fmt.Sprintf("guild/%s/track/%s", guildId, id.String())
	_, err := h.request(http.MethodDelete, url, nil, nil, &resData)
	if err != nil {
		return nil, err
	}

	return &resData, nil
}

func (h *HttpClient) request(
	method, url string,
	query url.Values,
	reqData, resData any,
) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	var reqBodyR io.Reader
	if reqData != nil {
		buf, err := json.Marshal(reqData)
		if err != nil {
			return false, errors.Unexpected(
				"track client: marshal request body: " + err.Error(),
			)
		}
		reqBodyR = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, h.baseUrl+url, reqBodyR)
	if err != nil {
		return false, errors.Unexpected(
			"track client: build request: " + err.Error(),
		)
	}

	if h.passwd != "" {
		req.Header.Add("Authorization", "passwd:"+h.passwd)
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	res, err := h.c.Do(req)
	if err != nil {
		return false, errors.Unexpected(
			"track client: do request: " + err.Error(),
		)
	}
	defer res.Body.Close()

	resBuf, err := io.ReadAll(res.Body)
	if err != nil {
		return false, errors.Unexpected(
			"track client: read response body: " + err.Error(),
		)
	}

	if res.StatusCode != http.StatusOK {
		var errb errBody
		if err = json.Unmarshal(resBuf, &errb); err != nil {
			return false, errors.Unexpected(
				"track client: unmarshal error response: " + err.Error(),
			)
		}
		if err = h.validate.Struct(&errb); err != nil {
			return false, errors.Unexpected(
				"track client: validate error response: " + err.Error(),
			)
		}

		if e := codeToErr(uint8(errb.ErrorCode)); e != nil {
			return false, e
		}

		return false, errors.Unexpected(
			"track client: error response: " + errb.Error,
		)
	}

	if resData != nil {
		dataB := dataBody{Data: resData}
		if err = json.Unmarshal(resBuf, &dataB); err != nil {
			return false, errors.Unexpected(
				"track client: unmarshal response: " + err.Error(),
			)
		}
		if err = h.validate.Struct(&dataB); err != nil {
			return false, errors.Unexpected(
				"track client: validate response: " + err.Error(),
			)
		}

		return dataB.Changed, nil
	}

	return false, nil
}
