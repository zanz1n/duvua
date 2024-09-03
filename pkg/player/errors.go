package player

import "github.com/zanz1n/duvua-bot/internal/errors"

var (
	ErrTooMuchTimePaused     = errors.New("muito tempo de pausa")
	ErrVoiceConnectionClosed = errors.New("desconectado do canal de voz")
	ErrTrackSearchFailed     = errors.New("não foi possível encontrar a música")

	ErrTrackNotFoundInQueue = errors.New("não foi possível achar a música na fila")
	ErrNoActivePlayer       = errors.New("o servidor não tem um player ativo")
)

const (
	ErrAnyCode uint8 = iota
	ErrTooMuchTimePausedCode
	ErrVoiceConnectionClosedCode
	ErrTrackSearchFailedCode
	ErrTrackNotFoundInQueueCode
	ErrNoActivePlayerCode
)

func codeToErr(code uint8) error {
	switch code {
	case ErrTooMuchTimePausedCode:
		return ErrTooMuchTimePaused
	case ErrVoiceConnectionClosedCode:
		return ErrVoiceConnectionClosed
	case ErrTrackSearchFailedCode:
		return ErrTrackSearchFailed
	case ErrTrackNotFoundInQueueCode:
		return ErrTrackNotFoundInQueue
	case ErrNoActivePlayerCode:
		return ErrNoActivePlayer
	default:
		return nil
	}
}
