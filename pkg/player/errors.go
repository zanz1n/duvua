package player

import "github.com/zanz1n/duvua/internal/errors"

var (
	ErrTooMuchTimePaused     = errors.New("muito tempo de pausa")
	ErrVoiceConnectionClosed = errors.New("desconectado do canal de voz")
	ErrTrackSearchFailed     = errors.New("não foi possível encontrar a música")
	ErrTrackSearchInvalidUrl = errors.New("a url fornecida é inválida")
	ErrTrackSearchUnsuported = errors.New("a url fornecida é de uma plataforma não suportada")

	ErrTrackNotFoundInQueue = errors.New("não foi possível achar a música na fila")
	ErrNoActivePlayer       = errors.New("o servidor não tem um player ativo")

	ErrSpotifyPlaylistsNotSupported = errors.New("playlists e álbuns do spotify não são suportados")
)

const (
	ErrAnyCode uint8 = iota
	ErrTooMuchTimePausedCode
	ErrVoiceConnectionClosedCode
	ErrTrackSearchFailedCode
	ErrTrackSearchInvalidUrlCode
	ErrTrackSearchUnsuportedCode
	ErrTrackNotFoundInQueueCode
	ErrNoActivePlayerCode
	ErrSpotifyPlaylistsNotSupportedCode
)

func codeToErr(code uint8) error {
	switch code {
	case ErrTooMuchTimePausedCode:
		return ErrTooMuchTimePaused
	case ErrVoiceConnectionClosedCode:
		return ErrVoiceConnectionClosed
	case ErrTrackSearchFailedCode:
		return ErrTrackSearchFailed
	case ErrTrackSearchInvalidUrlCode:
		return ErrTrackSearchInvalidUrl
	case ErrTrackNotFoundInQueueCode:
		return ErrTrackNotFoundInQueue
	case ErrNoActivePlayerCode:
		return ErrNoActivePlayer
	case ErrSpotifyPlaylistsNotSupportedCode:
		return ErrSpotifyPlaylistsNotSupported
	default:
		return nil
	}
}
