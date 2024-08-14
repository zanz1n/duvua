package ticket

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrInvalidChannelId = errors.Unexpected("ticket: channelId is not a valid int64")
	ErrInvalidUserId    = errors.Unexpected("ticket: userId is not a valid int64")
	ErrInvalidGuildId   = errors.Unexpected("ticket: guildId is not a valid int64")

	ErrInvalidSlug = errors.Unexpected("ticket: the slug is invalid")
)
