package welcome

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrInvalidId        = errors.Unexpected("welcome: id is not a valid int64")
	ErrInvalidChannelId = errors.Unexpected("welcome: channelId is not a valid int64")
)
