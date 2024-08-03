package welcome

import (
	"context"
	"database/sql"
	"time"
)

func NewPostgresWelcomeRepository(pool *sql.DB) WelcomeRepository {
	return &PostgresWelcomeRepository{
		pool: pool,
	}
}

type PostgresWelcomeRepository struct {
	pool      *sql.DB
	opTimeout time.Duration
}

// Create implements WelcomeRepository.
func (r *PostgresWelcomeRepository) Create(
	id string,
	enabled bool,
	channelId *string,
	message *string,
	kind *WelcomeType,
) (*Welcome, error) {
	const Query string = "INSERT INTO welcome " +
		"(id, enabled, channel_id, message, kind) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"RETURNING id, created_at, updated_at, enabled, channel_id, message, kind"

	data, err := newCreateChannelData(id, enabled, channelId, message, kind)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	var w postgresWelcome
	row := r.pool.QueryRowContext(ctx, Query,
		data.ID, data.Enabled, data.ChannelId, data.Message, data.Kind,
	)

	err = row.Scan(
		&w.ID,
		&w.CreatedAt,
		&w.UpdatedAt,
		&w.Enabled,
		&w.ChannelId,
		&w.Message,
		&w.Kind,
	)
	if err != nil {
		return nil, err
	}

	w2 := w.Into()
	return &w2, nil
}

// CreateDefault implements WelcomeRepository.
func (r *PostgresWelcomeRepository) CreateDefault(id string) (*Welcome, error) {
	const Query string = "INSERT INTO welcome (id) VALUES ($1) " +
		"RETURNING id, created_at, updated_at, enabled, channel_id, message, kind"

	id2, err := atoi(id)
	if err != nil {
		return nil, ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()
	row := r.pool.QueryRowContext(ctx, Query, id2)

	var w postgresWelcome
	err = row.Scan(
		&w.ID,
		&w.CreatedAt,
		&w.UpdatedAt,
		&w.Enabled,
		&w.ChannelId,
		&w.Message,
		&w.Kind,
	)
	if err != nil {
		return nil, err
	}

	w2 := w.Into()
	return &w2, nil
}

// Exists implements WelcomeRepository.
func (r *PostgresWelcomeRepository) Exists(id string) (bool, error) {
	const Query string = "SELECT id FROM welcome WHERE id = $1"

	id2, err := atoi(id)
	if err != nil {
		return false, ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	res, err := r.pool.QueryContext(ctx, Query, id2)
	if err != nil {
		return false, err
	}

	exists := res.Next()

	return exists, nil
}

// GetById implements WelcomeRepository.
func (r *PostgresWelcomeRepository) GetById(id string) (*Welcome, error) {
	const Query string = "SELECT " +
		"id, created_at, updated_at, enabled, channel_id, message, kind " +
		"FROM welcome WHERE id = $1"

	id2, err := atoi(id)
	if err != nil {
		return nil, ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()
	row := r.pool.QueryRowContext(ctx, Query, id2)

	var w postgresWelcome
	err = row.Scan(
		&w.ID,
		&w.CreatedAt,
		&w.UpdatedAt,
		&w.Enabled,
		&w.ChannelId,
		&w.Message,
		&w.Kind,
	)
	if err != nil {
		return nil, err
	}

	w2 := w.Into()
	return &w2, nil
}

// UpdateChannelId implements WelcomeRepository.
func (r *PostgresWelcomeRepository) UpdateChannelId(id string, channelId string) error {
	const Query string = "UPDATE welcome " +
		"SET channel_id = $1, updated_at = CURRENT_TIMESTAMP " +
		"WHERE id = $2"

	id2, err := atoi(id)
	if err != nil {
		return ErrInvalidId
	}

	channelId2, err := atoi(channelId)
	if err != nil {
		return ErrInvalidChannelId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	_, err = r.pool.ExecContext(ctx, Query, channelId2, id2)
	return err
}

// UpdateEnabled implements WelcomeRepository.
func (r *PostgresWelcomeRepository) UpdateEnabled(id string, enabled bool) error {
	const Query string = "UPDATE welcome " +
		"SET enabled = $1, updated_at = CURRENT_TIMESTAMP " +
		"WHERE id = $2"

	id2, err := atoi(id)
	if err != nil {
		return ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	_, err = r.pool.ExecContext(ctx, Query, enabled, id2)
	return err
}

// UpdateMessage implements WelcomeRepository.
func (r *PostgresWelcomeRepository) UpdateMessage(id string, message string, kind WelcomeType) error {
	const Query string = "UPDATE welcome SET " +
		"message = $1, kind = $2, updated_at = CURRENT_TIMESTAMP " +
		"WHERE id = $3"

	id2, err := atoi(id)
	if err != nil {
		return ErrInvalidId
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opTimeout)
	defer cancel()

	_, err = r.pool.ExecContext(ctx, Query, message, kind, id2)
	return err
}
