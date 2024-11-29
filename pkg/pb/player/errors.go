package player

import (
	"strconv"
	"strings"

	"github.com/zanz1n/duvua/internal/errors"
)

var (
	errTooMuchTimePaused     = errors.New("muito tempo de pausa")
	errVoiceConnectionClosed = errors.New("desconectado do canal de voz")
	errTrackSearchFailed     = errors.New("não foi possível encontrar a música")
	errTrackSearchInvalidUrl = errors.New("a url fornecida é inválida")
	errTrackSearchUnsuported = errors.New("a url fornecida é de uma plataforma não suportada")

	errTrackNotFoundInQueue = errors.New("não foi possível achar a música na fila")
	errNoActivePlayer       = errors.New("o servidor não tem um player ativo")

	errSpotifyPlaylistsNotSupported = errors.New("playlists e álbuns do spotify não são suportados")
)

func ConvertError(msg string) error {
	num, _, ok := strings.Cut(msg, ": ")
	if !ok {
		return nil
	}

	i, err2 := strconv.Atoi(num)
	if err2 != nil {
		return nil
	}

	newErr := CodeToErr(PlayerError(i))
	if newErr == nil {
		return nil
	}

	return newErr
}

func CodeToErr(code PlayerError) error {
	switch code {
	case PlayerError_ErrTooMuchTimePaused:
		return errTooMuchTimePaused
	case PlayerError_ErrVoiceConnectionClosed:
		return errVoiceConnectionClosed
	case PlayerError_ErrTrackSearchFailed:
		return errTrackSearchFailed
	case PlayerError_ErrTrackSearchInvalidUrl:
		return errTrackSearchInvalidUrl
	case PlayerError_ErrTrackSearchUnsuported:
		return errTrackSearchUnsuported
	case PlayerError_ErrTrackNotFoundInQueue:
		return errTrackNotFoundInQueue
	case PlayerError_ErrNoActivePlayer:
		return errNoActivePlayer
	case PlayerError_ErrSpotifyPlaylistsNotSupported:
		return errSpotifyPlaylistsNotSupported
	default:
		return nil
	}
}
