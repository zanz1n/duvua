package ticket

import (
	"context"
	"database/sql"
	"time"
)

var _ TicketConfigRepository = &PgTicketConfigRepository{}

func NewPgTicketConfigRepository(db *sql.DB) *PgTicketConfigRepository {
	return &PgTicketConfigRepository{
		db:        db,
		opTimeout: 2 * time.Second,
	}
}

type PgTicketConfigRepository struct {
	db        *sql.DB
	opTimeout time.Duration
}

// Create implements TicketConfigRepository.
func (r *PgTicketConfigRepository) Create(data TicketConfigCreateData) (*TicketConfig, error) {
	const Query = "INSERT INTO ticket_config " +
		"(guild_id, enabled, allow_multiple, channel_name, " +
		"channel_category_id, logs_channel_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6) RETURNING " +
		"guild_id, created_at, updated_at, enabled, allow_multiple, " +
		"channel_name, channel_category_id, logs_channel_id"

	pgdata, err := newPgCreateTicketConfigData(data)
	if err != nil {
		return nil, err
	}

	return r.fetch(
		Query,
		pgdata.GuildId,
		pgdata.Enabled,
		pgdata.AllowMultiple,
		pgdata.ChannelName,
		pgdata.ChannelCategoryId,
		pgdata.LogsChannelId,
	)
}

// CreateDefault implements TicketConfigRepository.
func (r *PgTicketConfigRepository) CreateDefault(guildId string) (*TicketConfig, error) {
	const Query = "INSERT INTO ticket_config (guild_id) VALUES ($1) RETURNING " +
		"guild_id, created_at, updated_at, enabled, allow_multiple, " +
		"channel_name, channel_category_id, logs_channel_id"

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	return r.fetch(Query, guildId2)
}

// ExistsByGuildId implements TicketConfigRepository.
// func (r *PgTicketConfigRepository) ExistsByGuildId(guildId string) (bool, error) {
// 	panic("unimplemented")
// }

// GetByGuildId implements TicketConfigRepository.
func (r *PgTicketConfigRepository) GetByGuildId(guildId string) (*TicketConfig, error) {
	const Query = "SELECT guild_id, created_at, updated_at, enabled, " +
		"allow_multiple, channel_name, channel_category_id, logs_channel_id " +
		"FROM ticket_config WHERE guild_id = $1"

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	return r.fetch(Query, guildId2)
}

// UpdateAllowMultiple implements TicketConfigRepository.
func (r *PgTicketConfigRepository) UpdateAllowMultiple(guildId string, allowMultiple bool) error {
	const Query = "UPDATE ticket_config SET allow_multiple = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	return r.exec(Query, allowMultiple, guildId2)
}

// UpdateChannelCategory implements TicketConfigRepository.
func (r *PgTicketConfigRepository) UpdateChannelCategory(guildId string, channelCategoryId string) error {
	const Query = "UPDATE ticket_config SET channel_category_id = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	channelCategoryId2 := sql.NullInt64{Valid: false}
	if channelCategoryId != "" {
		if channelCategoryId2.Int64, err = atoi(channelCategoryId); err != nil {
			return ErrInvalidChannelId
		}
		channelCategoryId2.Valid = true
	}

	return r.exec(Query, channelCategoryId2, guildId2)
}

// UpdateChannelName implements TicketConfigRepository.
func (r *PgTicketConfigRepository) UpdateChannelName(guildId string, channelName string) error {
	const Query = "UPDATE ticket_config SET channel_name = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	channelName2 := sql.NullString{Valid: false}
	if channelName != "" {
		channelName2.String = channelName
	}

	return r.exec(Query, channelName2, guildId2)
}

// UpdateEnabled implements TicketConfigRepository.
func (r *PgTicketConfigRepository) UpdateEnabled(guildId string, enabled bool) error {
	const Query = "UPDATE ticket_config SET enabled = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	return r.exec(Query, enabled, guildId2)
}

// UpdateLogsChannel implements TicketConfigRepository.
func (r *PgTicketConfigRepository) UpdateLogsChannel(guildId string, logsChannelId string) error {
	const Query = "UPDATE ticket_config SET logs_channel_id = $1 WHERE guild_id = $2"

	guildId2, err := atoi(guildId)
	if err != nil {
		return ErrInvalidGuildId
	}

	logsChannelId2 := sql.NullInt64{Valid: false}
	if logsChannelId != "" {
		if logsChannelId2.Int64, err = atoi(logsChannelId); err != nil {
			return ErrInvalidChannelId
		}
		logsChannelId2.Valid = true
	}

	return r.exec(Query, logsChannelId2, guildId2)
}

func (r *PgTicketConfigRepository) exec(query string, args ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PgTicketConfigRepository) fetch(query string, args ...any) (*TicketConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	var t pgTicketConfig
	row := r.db.QueryRowContext(ctx, query, args...)

	err := row.Scan(
		&t.GuildId,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Enabled,
		&t.AllowMultiple,
		&t.ChannelName,
		&t.ChannelCategoryId,
		&t.LogsChannelId,
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
