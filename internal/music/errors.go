package music

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrInvalidRolelId = errors.Unexpected("music: roleId is not a valid int64")
	ErrInvalidGuildId = errors.Unexpected("music: guildId is not a valid int64")
)
