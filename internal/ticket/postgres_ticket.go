package ticket

import (
	"context"
	"database/sql"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var _ TicketRepository = &PgTicketRepository{}

func NewPgTicketRepository(db *sql.DB) *PgTicketRepository {
	return &PgTicketRepository{
		db:         db,
		opTimeout:  2 * time.Second,
		nanoidSize: TicketSlugLength,
	}
}

type PgTicketRepository struct {
	db         *sql.DB
	opTimeout  time.Duration
	nanoidSize int
}

func (r *PgTicketRepository) genSlug() string {
	return gonanoid.MustGenerate(TicketSlugAlphabet, r.nanoidSize)
}

// Create implements TicketRepository.
func (r *PgTicketRepository) Create(channelId, userId, guildId string) (*Ticket, error) {
	const Query string = "INSERT INTO ticket " +
		"(slug, channel_id, user_id, guild_id) " +
		"VALUES ($1, $2, $3, $4) " +
		"RETURNING slug, created_at, channel_id, user_id, guild_id"

	data, err := newPgCreateTicketData(r.genSlug(), channelId, userId, guildId)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	row := r.db.QueryRowContext(ctx, Query,
		data.Slug, data.ChannelId, data.UserId, data.GuildId,
	)

	var t pgTicket
	err = row.Scan(
		&t.Slug,
		&t.CreatedAt,
		&t.ChannelId,
		&t.UserId,
		&t.GuildId,
	)
	if err != nil {
		return nil, err
	}

	t2 := t.Into()
	return &t2, nil
}

// DeleteByChannelId implements TicketRepository.
func (r *PgTicketRepository) DeleteByChannelId(channelId string) (*Ticket, error) {
	const Query string = "UPDATE ticket SET deleted_at = CURRENT_TIMESTAMP " +
		"WHERE channel_id = $1 AND deleted_at = NULL " +
		"RETURNING slug, created_at, channel_id, user_id, guild_id"

	channelId2, err := atoi(channelId)
	if err != nil {
		return nil, ErrInvalidChannelId
	}

	return r.fetch(Query, channelId2)
}

// DeleteByMember implements TicketRepository.
func (r *PgTicketRepository) DeleteByMember(guildId, userId string) ([]Ticket, error) {
	const Query string = "UPDATE ticket SET deleted_at = CURRENT_TIMESTAMP " +
		"WHERE guild_id = $1 AND user_id = $2 AND deleted_at = NULL " +
		"RETURNING slug, created_at, channel_id, user_id, guild_id"

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}
	userId2, err := atoi(userId)
	if err != nil {
		return nil, ErrInvalidUserId
	}

	return r.fetchMany(Query, guildId2, userId2)
}

// DeleteBySlug implements TicketRepository.
func (r *PgTicketRepository) DeleteBySlug(slug string) (*Ticket, error) {
	const Query string = "UPDATE ticket SET deleted_at = CURRENT_TIMESTAMP " +
		"WHERE slug = $1 AND deleted_at = NULL " +
		"RETURNING slug, created_at, channel_id, user_id, guild_id"

	if len(slug) != r.nanoidSize {
		return nil, ErrInvalidSlug
	}

	return r.fetch(Query, slug)
}

// GetByChannelId implements TicketRepository.
func (r *PgTicketRepository) GetByChannelId(channelId string) (*Ticket, error) {
	const Query string = "SELECT slug, created_at, channel_id, user_id, guild_id " +
		"FROM ticket WHERE channel_id = $1 AND deleted_at = NULL"

	channelId2, err := atoi(channelId)
	if err != nil {
		return nil, ErrInvalidChannelId
	}

	return r.fetch(Query, channelId2)
}

// GetByMember implements TicketRepository.
func (r *PgTicketRepository) GetByMember(guildId, userId string) ([]Ticket, error) {
	const Query string = "SELECT slug, created_at, channel_id, user_id, guild_id " +
		"FROM ticket WHERE guild_id = $1 AND user_id = $2 AND deleted_at = NULL"

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}
	userId2, err := atoi(userId)
	if err != nil {
		return nil, ErrInvalidUserId
	}

	return r.fetchMany(Query, guildId2, userId2)
}

// GetBySlug implements TicketRepository.
func (r *PgTicketRepository) GetBySlug(slug string) (*Ticket, error) {
	const Query string = "SELECT slug, created_at, channel_id, user_id, guild_id " +
		"FROM ticket WHERE slug = $1 AND deleted_at = NULL"

	if len(slug) != r.nanoidSize {
		return nil, ErrInvalidSlug
	}

	return r.fetch(Query, slug)
}

func (r *PgTicketRepository) fetchMany(query string, args ...any) ([]Ticket, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []Ticket{}

	for rows.Next() {
		var t Ticket

		err = rows.Scan(
			&t.Slug,
			&t.CreatedAt,
			&t.ChannelId,
			&t.UserId,
			&t.GuildId,
		)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (r *PgTicketRepository) fetch(query string, args ...any) (*Ticket, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	var t pgTicket
	row := r.db.QueryRowContext(ctx, query, args...)

	err := row.Scan(
		&t.Slug,
		&t.CreatedAt,
		&t.ChannelId,
		&t.UserId,
		&t.GuildId,
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
