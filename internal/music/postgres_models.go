package music

import (
	"database/sql"
	"strconv"
	"time"
)

type pgMusicConfig struct {
	GuildId     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Enabled     bool
	PlayMode    MusicPermission
	ControlMode MusicPermission

	DjRole sql.NullInt64
}

func (mc pgMusicConfig) Into() MusicConfig {
	djRole := ""
	if mc.DjRole.Valid {
		djRole = itoa(mc.DjRole.Int64)
	}

	return MusicConfig{
		GuildId:     itoa(mc.GuildId),
		CreatedAt:   mc.CreatedAt,
		UpdatedAt:   mc.UpdatedAt,
		Enabled:     mc.Enabled,
		PlayMode:    mc.PlayMode,
		ControlMode: mc.ControlMode,
		DjRole:      djRole,
	}
}

type pgMusicConfigCreateData struct {
	GuildId     int64
	Enabled     bool
	PlayMode    MusicPermission
	ControlMode MusicPermission
	DjRole      sql.NullInt64
}

func newPgMusicConfigCreateData(data MusicConfigCreateData) (*pgMusicConfigCreateData, error) {
	guildId, err := atoi(data.GuildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	djRole := sql.NullInt64{}
	if data.DjRole != "" {
		djRole.Int64, err = atoi(data.DjRole)
		if err != nil {
			return nil, ErrInvalidRolelId
		}
		djRole.Valid = true
	}

	pgData := pgMusicConfigCreateData{
		GuildId:     guildId,
		Enabled:     data.Enabled,
		PlayMode:    data.PlayMode,
		ControlMode: data.ControlMode,
		DjRole:      djRole,
	}

	if data.PlayMode == "" {
		pgData.PlayMode = DefaultConfigPlayMode
	}
	if data.ControlMode == "" {
		pgData.ControlMode = DefaultConfigControlMode
	}

	return &pgData, nil
}

func atoi(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
