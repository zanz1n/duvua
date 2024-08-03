package welcome

import (
	"database/sql"
	"time"

	"github.com/zanz1n/duvua-bot/internal/errors"
)

type postgresWelcome struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Enabled   bool
	ChannelId sql.NullInt64
	Message   string
	Kind      WelcomeType
}

func (w postgresWelcome) Into() Welcome {
	w2 := Welcome{
		ID:        itoa(w.ID),
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		Enabled:   w.Enabled,
		ChannelId: nil,
		Message:   w.Message,
		Kind:      w.Kind,
	}

	if w.ChannelId.Valid {
		channelId := itoa(w.ChannelId.Int64)
		w2.ChannelId = &channelId
	}

	return w2
}

type createChannelData struct {
	ID        int64
	Enabled   bool
	ChannelId sql.NullInt64
	Message   string
	Kind      WelcomeType
}

func newCreateChannelData(
	id string,
	enabled bool,
	channelId *string,
	message *string,
	kind *WelcomeType,
) (*createChannelData, error) {
	id2, err := atoi(id)
	if err != nil {
		return nil, ErrInvalidId
	}

	channelId2 := sql.NullInt64{Valid: false}

	if channelId != nil {
		if channelId2.Int64, err = atoi(*channelId); err != nil {
			return nil, ErrInvalidId
		} else {
			channelId2.Valid = true
		}
	}

	data := createChannelData{
		ID:        id2,
		Enabled:   enabled,
		ChannelId: channelId2,
	}

	if message != nil {
		data.Message = *message
	} else {
		data.Message = DefaultMessage
	}

	if kind != nil {
		data.Kind = *kind
	} else {
		data.Kind = DefaultKind
	}

	return &data, nil
}

func (e *WelcomeType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = WelcomeType(s)
	case string:
		*e = WelcomeType(s)
	default:
		return errors.Unexpectedf("unsupported scan type for QuestionStyle: %T", src)
	}
	return nil
}
