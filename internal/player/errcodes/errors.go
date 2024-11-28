package errcodes

import (
	"github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrTooMuchTimePaused = status.Errorf(
		codes.Canceled,
		"%d: too much time paused",
		player.PlayerError_ErrTooMuchTimePaused,
	)
	ErrVoiceConnectionClosed = status.Errorf(
		codes.Aborted,
		"%d: voice connection closed",
		player.PlayerError_ErrVoiceConnectionClosed,
	)
	ErrTrackSearchFailed = status.Errorf(
		codes.NotFound,
		"%d: couldn't find track",
		player.PlayerError_ErrTrackSearchFailed,
	)
	ErrTrackSearchInvalidUrl = status.Errorf(
		codes.InvalidArgument,
		"%d: the provided search url is invalid",
		player.PlayerError_ErrTrackSearchInvalidUrl,
	)
	ErrTrackSearchUnsuported = status.Errorf(
		codes.Unimplemented,
		"%d: the url is of an unsuported platform",
		player.PlayerError_ErrTrackSearchUnsuported,
	)

	ErrTrackNotFoundInQueue = status.Errorf(
		codes.NotFound,
		"%d: the track could not be found in the queue",
		player.PlayerError_ErrTrackNotFoundInQueue,
	)
	ErrNoActivePlayer = status.Errorf(
		codes.FailedPrecondition,
		"%d: there is not an active player",
		player.PlayerError_ErrNoActivePlayer,
	)

	ErrSpotifyPlaylistsNotSupported = status.Errorf(
		codes.Unimplemented,
		"%d: spotify playlist are not supported",
		player.PlayerError_ErrSpotifyPlaylistsNotSupported,
	)
)

func ErrToErrCode(err error) player.PlayerError {
	switch err {
	case ErrTooMuchTimePaused:
		return player.PlayerError_ErrTooMuchTimePaused
	case ErrVoiceConnectionClosed:
		return player.PlayerError_ErrVoiceConnectionClosed
	case ErrTrackSearchFailed:
		return player.PlayerError_ErrTrackSearchFailed
	case ErrTrackSearchInvalidUrl:
		return player.PlayerError_ErrTrackSearchInvalidUrl
	case ErrTrackSearchUnsuported:
		return player.PlayerError_ErrTrackSearchUnsuported
	case ErrTrackNotFoundInQueue:
		return player.PlayerError_ErrTrackNotFoundInQueue
	case ErrNoActivePlayer:
		return player.PlayerError_ErrNoActivePlayer
	case ErrSpotifyPlaylistsNotSupported:
		return player.PlayerError_ErrSpotifyPlaylistsNotSupported
	default:
		return player.PlayerError_ErrAny
	}
}
