package welcome

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrInvalidId        = errors.Unexpected("id is not a valid int64")
	ErrInvalidChannelId = errors.Unexpected("channelId is not a valid int64")
)
