package lang

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zanz1n/duvua/internal/errors"
)

var errParseInvalidData = errors.Unexpected("translator parse: invalid data")

var _ Translator = &GoogleTranslatorApi{}

func NewGoogleTranslatorApi(client *http.Client) *GoogleTranslatorApi {
	if client == nil {
		client = http.DefaultClient
	}

	return &GoogleTranslatorApi{
		Client:  client,
		Timeout: 2 * time.Second,
	}
}

type GoogleTranslatorApi struct {
	Client  *http.Client
	Timeout time.Duration
}

// Translate implements Translator.
func (t *GoogleTranslatorApi) Translate(srcl, dstl Language, source string) (string, error) {
	bodyBuf, err := t.requestTranslation(srcl, dstl, source)
	if err != nil {
		return "", err
	}

	var body []any
	if err = json.Unmarshal(bodyBuf, &body); err != nil {
		return "", errors.Unexpected("translator parse: " + err.Error())
	} else if len(body) == 0 {
		return "", errParseInvalidData
	}

	inner, ok := body[0].([]any)
	if !ok {
		return "", errParseInvalidData
	}

	text := []string{}
	for _, slice := range inner {
		s, ok := slice.([]any)
		if !ok {
			return "", errParseInvalidData
		} else if len(s) == 0 {
			return "", errParseInvalidData
		}

		if t, ok := s[0].(string); ok {
			text = append(text, t)
		}
	}
	s := strings.Join(text, "")

	return s, nil
}

func (t *GoogleTranslatorApi) requestTranslation(sl, dl Language, source string) ([]byte, error) {
	url := "https://translate.googleapis.com/translate_a/single"

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.Unexpected("translator fetch: " + err.Error())
	}

	values := req.URL.Query()
	values.Add("client", "gtx")
	values.Add("sl", sl.String())
	values.Add("tl", dl.String())
	values.Add("dt", "t")
	values.Add("q", source)
	req.URL.RawQuery = values.Encode()

	res, err := t.Client.Do(req)
	if err != nil {
		return nil, errors.Unexpected("translator fetch: " + err.Error())
	}
	defer res.Body.Close()

	bodyBuf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Unexpected(
			"translator fetch: failed to read response body: " + err.Error(),
		)
	}
	return bodyBuf, nil
}
