package ticket

type TicketRepository interface {
	GenerateSlug() string

	// The returned Ticket clould be nil
	GetBySlug(slug string) (*Ticket, error)
	// The returned Ticket clould be nil
	GetByChannelId(channelId string) (*Ticket, error)
	// The returned Tickets clould be nil
	GetByMember(guildId, userId string) ([]Ticket, error)

	// The returned Ticket must not be nil if err != nil
	Create(slug, channelId, userId, guildId string) (*Ticket, error)

	// The returned ticket clould be nil
	DeleteBySlug(slug string) (*Ticket, error)
	// The returned Ticket clould be nil
	DeleteByChannelId(channelId string) (*Ticket, error)
	// The returned Tickets clould be nil
	DeleteByMember(guildId, userId string) ([]Ticket, error)
}

type TicketConfigRepository interface {
	// The returned TicketConfig clould be nil
	GetByGuildId(guildId string) (*TicketConfig, error)
	// ExistsByGuildId(guildId string) (bool, error)

	// The returned TicketConfig must not be nil if err != nil
	Create(data TicketConfigCreateData) (*TicketConfig, error)
	// The returned TicketConfig must not be nil if err != nil
	CreateDefault(guildId string) (*TicketConfig, error)

	UpdateEnabled(guildId string, enabled bool) error
	UpdateAllowMultiple(guildId string, allowMultiple bool) error
	UpdateChannelName(guildId string, channelName string) error
	UpdateChannelCategory(guildId string, channelCategoryId string) error
	UpdateLogsChannel(guildId string, logsChannelId string) error
}
