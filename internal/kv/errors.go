package kv

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrNotStringValue = errors.Unexpected("stored value is not of string type")
	ErrMismatchType   = errors.Unexpected("the type of v mismatches the stored value type")

	ErrFilePathNotProvided = errors.Unexpected("the file path was not provided in service instantiation")
)
