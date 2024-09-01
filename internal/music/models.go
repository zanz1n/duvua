package music

import "time"

const (
	DefaultConfigEnabled bool = true

	DefaultConfigPlayMode    = MusicPermissionAll
	DefaultConfigControlMode = MusicPermissionDJ
)

type MusicPermission string

const (
	MusicPermissionAll MusicPermission = "All"
	MusicPermissionDJ  MusicPermission = "DJ"
	MusicPermissionAdm MusicPermission = "Adm"
)

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
