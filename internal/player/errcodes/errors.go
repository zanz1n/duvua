package errcodes

import (
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/pkg/player"
)

var (
	ErrTooMuchTimePaused     = errors.New("too much time paused")
	ErrVoiceConnectionClosed = errors.New("voice connection closed")
	ErrTrackSearchFailed     = errors.New("couldn't find track")
	ErrTrackSearchInvalidUrl = errors.New("the provided search url is invalid")
	ErrTrackSearchUnsuported = errors.New("the url is of an unsuported platform")

	ErrTrackNotFoundInQueue = errors.New("the track could not be found in the queue")
	ErrNoActivePlayer       = errors.New("there is not an active player")

	ErrSpotifyPlaylistsNotSupported = errors.New("spotify playlist are not supported")
)

func ErrToErrCode(err error) uint8 {
	switch err {
	case ErrTooMuchTimePaused:
		return player.ErrTooMuchTimePausedCode
	case ErrVoiceConnectionClosed:
		return player.ErrVoiceConnectionClosedCode
	case ErrTrackSearchFailed:
		return player.ErrTrackSearchFailedCode
	case ErrTrackSearchInvalidUrl:
		return player.ErrTrackSearchInvalidUrlCode
	case ErrTrackSearchUnsuported:
		return player.ErrTrackSearchUnsuportedCode
	case ErrTrackNotFoundInQueue:
		return player.ErrTrackNotFoundInQueueCode
	case ErrNoActivePlayer:
		return player.ErrNoActivePlayerCode
	case ErrSpotifyPlaylistsNotSupported:
		return player.ErrSpotifyPlaylistsNotSupportedCode
	default:
		return player.ErrAnyCode
	}
}
