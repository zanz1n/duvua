package ticket

import "time"

const (
	DefaultConfigChannelName            string = "ticket-{{USER}}-{{ID}}"
	DefaultConfigShortChannelName       string = "t-{{USER}}-{{ID}}"
	DefaultConfigCategorizedChannelName string = "{{USER}}-{{ID}}"

	DefaultConfigEnabled       bool = false
	DefaultConfigAllowMultiple bool = true

	TicketSlugLength   int    = 8
	TicketSlugAlphabet string = "0123456789abcdefghijklmnopqrstuvwxyz"
)

type Ticket struct {
	// 8 Character long nanoid
	Slug      string
	CreatedAt time.Time
	// Not null
	ChannelId string
	// Not null
	UserId string
	// Not null
	GuildId string
}

type TicketConfig struct {
	GuildId       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Enabled       bool
	AllowMultiple bool
	// Nullable: coallessed to empty string
	//
	// The template to create the name of the ticket channel
	ChannelName string
	// Nullable: coallessed to empty string
	//
	// The discord channel category to put the ticket channels
	ChannelCategoryId string
	// Nullable: coallessed to empty string
	LogsChannelId string
}

type TicketConfigCreateData struct {
	GuildId           string
	Enabled           *bool
	AllowMultiple     *bool
	ChannelName       string
	ChannelCategoryId string
	LogsChannelId     string
}
