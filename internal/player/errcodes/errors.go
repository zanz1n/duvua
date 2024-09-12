package errcodes

import "github.com/zanz1n/duvua/internal/errors"

var (
	ErrTooMuchTimePaused     = errors.New("too much time paused")
	ErrVoiceConnectionClosed = errors.New("voice connection closed")
	ErrTrackSearchFailed     = errors.New("couldn't find track")

	ErrTrackNotFoundInQueue = errors.New("the track could not be found in the queue")
	ErrNoActivePlayer       = errors.New("there is not an active player")
)

const (
	ErrAnyCode uint8 = iota
	ErrTooMuchTimePausedCode
	ErrVoiceConnectionClosedCode
	ErrTrackSearchFailedCode
	ErrTrackNotFoundInQueueCode
	ErrNoActivePlayerCode
)

func ErrToErrCode(err error) uint8 {
	switch err {
	case ErrTooMuchTimePaused:
		return ErrTooMuchTimePausedCode
	case ErrVoiceConnectionClosed:
		return ErrVoiceConnectionClosedCode
	case ErrTrackSearchFailed:
		return ErrTrackSearchFailedCode
	case ErrTrackNotFoundInQueue:
		return ErrTrackNotFoundInQueueCode
	case ErrNoActivePlayer:
		return ErrNoActivePlayerCode
	default:
		return ErrAnyCode
	}
}
