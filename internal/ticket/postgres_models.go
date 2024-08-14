package ticket

import (
	"database/sql"
	"strconv"
	"time"
)

type pgTicket struct {
	ID        int32
	Slug      string
	CreatedAt time.Time
	ChannelId int64
	UserId    int64
	GuildId   int64
	DeletedAt sql.NullTime
}

func (t pgTicket) Into() Ticket {
	return Ticket{
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		ChannelId: itoa(t.ChannelId),
		UserId:    itoa(t.UserId),
		GuildId:   itoa(t.GuildId),
	}
}

type pgCreateTicketData struct {
	Slug      string
	ChannelId int64
	UserId    int64
	GuildId   int64
}

func newPgCreateTicketData(
	slug, channelId, userId, guildId string,
) (*pgCreateTicketData, error) {
	channelId2, err := atoi(channelId)
	if err != nil {
		return nil, ErrInvalidChannelId
	}

	userId2, err := atoi(userId)
	if err != nil {
		return nil, ErrInvalidUserId
	}

	guildId2, err := atoi(guildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	return &pgCreateTicketData{
		Slug:      slug,
		ChannelId: channelId2,
		UserId:    userId2,
		GuildId:   guildId2,
	}, nil
}

type pgTicketConfig struct {
	GuildId           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Enabled           bool
	AllowMultiple     bool
	ChannelName       sql.NullString
	ChannelCategoryId sql.NullInt64
	LogsChannelId     sql.NullInt64
}

func (t pgTicketConfig) Into() TicketConfig {
	channelCategoryId := ""
	if t.ChannelCategoryId.Valid {
		channelCategoryId = itoa(t.ChannelCategoryId.Int64)
	}
	logsChannelId := ""
	if t.LogsChannelId.Valid {
		logsChannelId = itoa(t.LogsChannelId.Int64)
	}
	channelName := ""
	if t.ChannelName.Valid {
		channelName = t.ChannelName.String
	}

	return TicketConfig{
		GuildId:           itoa(t.GuildId),
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
		Enabled:           t.Enabled,
		AllowMultiple:     t.AllowMultiple,
		ChannelCategoryId: channelCategoryId,
		LogsChannelId:     logsChannelId,
		ChannelName:       channelName,
	}
}

type pgCreateTicketConfigData struct {
	GuildId           int64
	Enabled           bool
	AllowMultiple     bool
	ChannelName       sql.NullString
	ChannelCategoryId sql.NullInt64
	LogsChannelId     sql.NullInt64
}

func newPgCreateTicketConfigData(
	data TicketConfigCreateData,
) (*pgCreateTicketConfigData, error) {
	var err error
	d := pgCreateTicketConfigData{
		Enabled:       DefaultConfigEnabled,
		AllowMultiple: DefaultConfigAllowMultiple,
	}

	d.GuildId, err = atoi(data.GuildId)
	if err != nil {
		return nil, ErrInvalidGuildId
	}

	if data.Enabled != nil {
		d.Enabled = *data.Enabled
	}
	if data.AllowMultiple != nil {
		d.AllowMultiple = *data.AllowMultiple
	}
	if data.ChannelName != "" {
		d.ChannelName = sql.NullString{
			Valid:  true,
			String: data.ChannelName,
		}
	}
	if data.ChannelCategoryId != "" {
		channelId, err := atoi(data.ChannelCategoryId)
		if err != nil {
			return nil, ErrInvalidChannelId
		}
		d.ChannelCategoryId = sql.NullInt64{
			Valid: true,
			Int64: channelId,
		}
	}
	if data.LogsChannelId != "" {
		channelId, err := atoi(data.ChannelCategoryId)
		if err != nil {
			return nil, ErrInvalidChannelId
		}
		d.LogsChannelId = sql.NullInt64{
			Valid: true,
			Int64: channelId,
		}
	}

	return &d, nil
}

func atoi(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
