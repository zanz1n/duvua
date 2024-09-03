package music

import (
	"context"
	"database/sql"
	"time"
)

var _ MusicConfigRepository = &PgMusicConfigRepository{}

func NewPgMusicConfigRepository(db *sql.DB) *PgMusicConfigRepository {
	return &PgMusicConfigRepository{
		db:        db,
		opTimeout: 2 * time.Second,
	}
}

type PgMusicConfigRepository struct {
	db        *sql.DB
	opTimeout time.Duration
}

// Create implements MusicConfigRepository.
func (r *PgMusicConfigRepository) Create(data MusicConfigCreateData) (*MusicConfig, error) {
	const Query = "INSERT INTO music_config (guild_id, enabled, play_mode, " +
		"control_mode, dj_role) VALUES ($1, $2, $3, $4, $5) RETURNING " +
		"guild_id, created_at, updated_at, enabled, play_mode, control_mode, dj_role"

	pgdata, err := newPgMusicConfigCreateData(data)
	if err != nil {
		return nil, err
	}

	return r.fetch(
		Query,
		pgdata.GuildId,
		pgdata.Enabled,
		pgdata.PlayMode,
		pgdata.ControlMode,
		pgdata.DjRole,
	)
}

// GetByGuildId implements MusicConfigRepository.
func (r *PgMusicConfigRepository) GetByGuildId(guildId string) (*MusicConfig, error) {
	const Query = "SELECT guild_id, created_at, updated_at, enabled, play_mode, " +
		"control_mode, dj_role FROM music_config WHERE guild_id = $1"

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	return r.fetch(Query, guildId2)
}

// GetOrDefault implements MusicConfigRepository.
func (r *PgMusicConfigRepository) GetOrDefault(guildId string) (*MusicConfig, error) {
	c, err := r.GetByGuildId(guildId)
	if err != nil {
		return nil, err
	}

	if c == nil {
		now := time.Now()
		c = &MusicConfig{
			GuildId:     guildId,
			CreatedAt:   now,
			UpdatedAt:   now,
			Enabled:     DefaultConfigEnabled,
			PlayMode:    DefaultConfigPlayMode,
			ControlMode: DefaultConfigControlMode,
			DjRole:      "",
		}
	}

	return c, nil
}

// UpdateControlMode implements MusicConfigRepository.
func (r *PgMusicConfigRepository) UpdateControlMode(
	guildId string,
	controlMode MusicPermission,
) error {
	const Query = "UPDATE music_config SET control_mode = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	return r.exec(Query, controlMode, guildId2)
}

// UpdateDjRole implements MusicConfigRepository.
func (r *PgMusicConfigRepository) UpdateDjRole(guildId string, djRole string) error {
	const Query = "UPDATE music_config SET dj_role = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	djRole2 := sql.NullInt64{}
	if djRole != "" {
		if djRole2.Int64, err = atoi(djRole); err != nil {
			return ErrInvalidRolelId
		}
		djRole2.Valid = true
	}

	return r.exec(Query, djRole2, guildId2)
}

// UpdateEnabled implements MusicConfigRepository.
func (r *PgMusicConfigRepository) UpdateEnabled(guildId string, enabled bool) error {
	const Query = "UPDATE music_config SET enabled = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	return r.exec(Query, enabled, guildId2)
}

// UpdatePlayMode implements MusicConfigRepository.
func (r *PgMusicConfigRepository) UpdatePlayMode(
	guildId string,
	playMode MusicPermission,
) error {
	const Query = "UPDATE music_config SET play_mode = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	return r.exec(Query, playMode, guildId2)
}

func (r *PgMusicConfigRepository) exec(query string, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PgMusicConfigRepository) fetch(query string, args ...any) (*MusicConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	var t pgMusicConfig
	row := r.db.QueryRowContext(ctx, query, args...)

	err := row.Scan(
		&t.GuildId,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Enabled,
		&t.PlayMode,
		&t.ControlMode,
		&t.DjRole,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	t2 := t.Into()
	return &t2, nil
}
