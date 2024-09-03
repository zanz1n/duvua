package musiccmds

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/music"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

func canPlay(m *discordgo.Member, cfg *music.MusicConfig) bool {
	switch cfg.PlayMode {
	case music.MusicPermissionAll:
		return true
	case music.MusicPermissionAdm:
		return utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator)
	case music.MusicPermissionDJ:
		return utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator) ||
			slices.Contains(m.Roles, cfg.DjRole)
	default:
		return false
	}
}

func canControl(m *discordgo.Member, cfg *music.MusicConfig) error {
	can := false
	switch cfg.ControlMode {
	case music.MusicPermissionAll:
		can = true
	case music.MusicPermissionAdm:
		can = utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator)
	case music.MusicPermissionDJ:
		can = utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator) ||
			slices.Contains(m.Roles, cfg.DjRole)
	}

	if !can {
		return errors.New("você não tem permissão para controlar a playlist no servidor")
	}
	return nil
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
