package welcome

type WelcomeRepository interface {
	// The returned welcome clould be nil
	GetById(id string) (*Welcome, error)
	Exists(id string) (bool, error)

	// The returned welcome must not be nil if err != nil
	Create(
		id string,
		enabled bool,
		channelId *string,
		message *string,
		kind *WelcomeType,
	) (*Welcome, error)
	// The returned welcome must not be nil if err != nil
	CreateDefault(id string) (*Welcome, error)

	UpdateEnabled(id string, enabled bool) error
	UpdateChannelId(id string, channelId string) error
	UpdateMessage(id string, message string, kind WelcomeType) error
}
