package music

import (
	"time"

	"github.com/zanz1n/duvua/internal/errors"
)

const (
	DefaultConfigEnabled bool = true

	DefaultConfigPlayMode    = MusicPermissionAll
	DefaultConfigControlMode = MusicPermissionDJ
)

type MusicPermission string

func ParseMusicPermission(s string) (MusicPermission, error) {
	v := MusicPermission(s)
	switch v {
	case MusicPermissionAll, MusicPermissionDJ, MusicPermissionAdm:
		return v, nil
	default:
		return "", errors.Newf("valor inválido `%s`", s)
	}
}

const (
	MusicPermissionAll MusicPermission = "All"
	MusicPermissionDJ  MusicPermission = "DJ"
	MusicPermissionAdm MusicPermission = "Adm"
)

func (m MusicPermission) StringEnUs() string {
	switch m {
	case MusicPermissionAll:
		return "Any member"
	case MusicPermissionDJ:
		return "DJ's only"
	case MusicPermissionAdm:
		return "Only administrators"
	default:
		return "Unknown"
	}
}

func (m MusicPermission) StringPtBr() string {
	switch m {
	case MusicPermissionAll:
		return "Quaisquer membros"
	case MusicPermissionDJ:
		return "Apenas DJs"
	case MusicPermissionAdm:
		return "Apenas administradores"
	default:
		return "Desconhecido"
	}
}

type MusicConfig struct {
	GuildId     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Enabled     bool
	PlayMode    MusicPermission
	ControlMode MusicPermission
	// Nullable: coallessed to empty string
	DjRole string
}

type MusicConfigCreateData struct {
	GuildId     string
	Enabled     bool
	PlayMode    MusicPermission
	ControlMode MusicPermission
	DjRole      string
}
