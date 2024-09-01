package music

type MusicConfigRepository interface {
	// The returned MusicConfig clould be nil
	GetByGuildId(guildId string) (*MusicConfig, error)

	// The returned MusicConfig must not be nil if err != nil
	Create(data MusicConfigCreateData) (*MusicConfig, error)

	UpdateEnabled(guildId string, enabled bool) error
	UpdatePlayMode(guildId string, playMode MusicPermission) error
	UpdateControlMode(guildId string, controlMode MusicPermission) error
	UpdateDjRole(guildId string, djRole string) error
}
