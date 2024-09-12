package musiccmds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/utils"
)

var musicadminCommandData = discordgo.ApplicationCommand{
	Name:        "musicadmin",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Comandos de configuração para músicas",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Configuration commands for music functionality",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "enable",
			Description: "Habilita a funcionalidade de músicas no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Enables the music functionality on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "disable",
			Description: "Disabilita a funcionalidade de músicas no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Disables the music functionality on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "allow-play",
			Description: "Define quem poderá tocar músicas no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Defines who will be able to play musics on the server",
			},
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "Quem poderá tocar músicas",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.EnglishUS: "Who will be able to play musics",
				},
				Required: true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name: music.MusicPermissionAll.StringPtBr() + " (padrão)",
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionAll.StringEnUs() +
								" (default)",
						},
						Value: music.MusicPermissionAll,
					},
					{
						Name: music.MusicPermissionDJ.StringPtBr(),
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionDJ.StringEnUs(),
						},
						Value: music.MusicPermissionDJ,
					},
					{
						Name: music.MusicPermissionAdm.StringPtBr(),
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionAdm.StringEnUs(),
						},
						Value: music.MusicPermissionAdm,
					},
				},
			}},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "allow-control",
			Description: "Define quem poderá controlar músicas no servidor (skip, stop, etc...)",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Defines who will be able to control musics on the server",
			},
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "Quem poderá controlar músicas",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.EnglishUS: "Who will be able to control musics",
				},
				Required: true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name: music.MusicPermissionAll.StringPtBr(),
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionAll.StringEnUs(),
						},
						Value: music.MusicPermissionAll,
					},
					{
						Name: music.MusicPermissionDJ.StringPtBr() + " (padrão)",
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionDJ.StringEnUs() +
								" (default)",
						},
						Value: music.MusicPermissionDJ,
					},
					{
						Name: music.MusicPermissionAdm.StringPtBr(),
						NameLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: music.MusicPermissionAdm.StringEnUs(),
						},
						Value: music.MusicPermissionAdm,
					},
				},
			}},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set-dj",
			Description: "Define um cargo para os DJs",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Defines a role for the DJ's",
			},
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "O cargo dos DJ's",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.EnglishUS: "The DJ's role",
				},
				Required: true,
			}},
		},
	},
}

func NewMusicAdminCommand(r music.MusicConfigRepository) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &musicadminCommandData,
		Category: manager.CommandCategoryConfig,
		Handler:  &MusicAdminCommand{r: r},
	}
}

type MusicAdminCommand struct {
	r music.MusicConfigRepository
}

func (c *MusicAdminCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}
	if !utils.HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
		return errors.New("você não tem permissão para usar esse comando")
	}

	subCommand, err := i.GetSubCommand()
	if err != nil {
		return err
	}

	switch subCommand.Name {
	case "enable":
		return c.handleEnable(s, i, true)

	case "disable":
		return c.handleEnable(s, i, false)

	case "allow-play":
		modeOpt, err := i.GetStringOption("mode", true)
		if err != nil {
			return err
		}

		mode, err := music.ParseMusicPermission(modeOpt)
		if err != nil {
			return errors.New("opção `mode` inválida: " + err.Error())
		}

		return c.handleAllowPlay(s, i, mode)

	case "allow-control":
		modeOpt, err := i.GetStringOption("mode", true)
		if err != nil {
			return err
		}

		mode, err := music.ParseMusicPermission(modeOpt)
		if err != nil {
			return errors.New("opção `mode` inválida: " + err.Error())
		}

		return c.handleAllowControl(s, i, mode)

	case "set-dj":
		opt, err := i.GetTypedOption("role", true, discordgo.ApplicationCommandOptionRole)
		if err != nil {
			return err
		}

		role := opt.RoleValue(s, i.GuildID)
		if role.Name == "" {
			return errors.New("opção `role` precisa ser um cargo válido")
		}

		return c.handleSetDJ(s, i, role)

	default:
		return errors.New("opção `sub-command` inválida")
	}
}

func (c *MusicAdminCommand) handleEnable(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	enable bool,
) error {
	cfg, err := c.r.GetByGuildId(i.GuildID)
	if err != nil {
		return err
	}

	beforeEnabled := music.DefaultConfigEnabled
	if cfg != nil {
		beforeEnabled = cfg.Enabled
		if cfg.Enabled != enable {
			if err = c.r.UpdateEnabled(cfg.GuildId, enable); err != nil {
				return err
			}
		}
	} else if music.DefaultConfigEnabled != enable {
		_, err = c.r.Create(music.MusicConfigCreateData{
			GuildId: i.GuildID,
			Enabled: enable,
		})
		if err != nil {
			return err
		}
	}

	msg := "As músicas"
	if beforeEnabled == enable {
		msg += " já estavam"
	} else {
		msg += " foram"
	}
	if enable {
		msg += " habilitadas"
	} else {
		msg += " desabilitadas"
	}

	return i.Replyf(s, msg)
}

func (c *MusicAdminCommand) handleAllowPlay(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	mode music.MusicPermission,
) error {
	cfg, err := c.r.GetByGuildId(i.GuildID)
	if err != nil {
		return err
	}

	changed := false
	if cfg != nil && cfg.PlayMode != mode {
		if err = c.r.UpdatePlayMode(i.GuildID, mode); err != nil {
			return err
		}
		changed = true
	} else if cfg == nil && mode != music.DefaultConfigPlayMode {
		_, err = c.r.Create(music.MusicConfigCreateData{
			GuildId:  i.GuildID,
			Enabled:  music.DefaultConfigEnabled,
			PlayMode: mode,
		})
		if err != nil {
			return err
		}
		changed = true
	}

	if changed {
		return i.Replyf(s,
			"Configuração atualizada: %s podem tocar músicas no servidor",
			mode.StringPtBr(),
		)
	} else {
		return i.Replyf(s,
			"Configuração não mudou: %s já podiam tocar músicas no servidor",
			mode.StringPtBr(),
		)
	}
}

func (c *MusicAdminCommand) handleAllowControl(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	mode music.MusicPermission,
) error {
	cfg, err := c.r.GetByGuildId(i.GuildID)
	if err != nil {
		return err
	}

	changed := false
	if cfg != nil && cfg.ControlMode != mode {
		if err = c.r.UpdateControlMode(i.GuildID, mode); err != nil {
			return err
		}
		changed = true
	} else if cfg == nil && mode != music.DefaultConfigControlMode {
		_, err = c.r.Create(music.MusicConfigCreateData{
			GuildId:     i.GuildID,
			Enabled:     music.DefaultConfigEnabled,
			ControlMode: mode,
		})
		if err != nil {
			return err
		}
		changed = true
	}

	if changed {
		return i.Replyf(s,
			"Configuração atualizada: %s podem controlar músicas no servidor",
			mode.StringPtBr(),
		)
	} else {
		return i.Replyf(s,
			"Configuração não mudou: %s já podiam controlar músicas no servidor",
			mode.StringPtBr(),
		)
	}
}

func (c *MusicAdminCommand) handleSetDJ(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	role *discordgo.Role,
) error {
	cfg, err := c.r.GetByGuildId(i.GuildID)
	if err != nil {
		return err
	}

	changed := false
	if cfg != nil && cfg.DjRole != role.ID {
		if err = c.r.UpdateDjRole(i.GuildID, role.ID); err != nil {
			return err
		}
		changed = true
	} else if cfg == nil && role.ID != "" {
		_, err = c.r.Create(music.MusicConfigCreateData{
			GuildId: i.GuildID,
			Enabled: music.DefaultConfigEnabled,
			DjRole:  role.ID,
		})
		if err != nil {
			return err
		}
		changed = true
	}

	if changed {
		return i.Replyf(s,
			"Configuração atualizada: o cargo dos DJs foi definido para `%s`",
			role.Name,
		)
	} else {
		return i.Replyf(s,
			"Configuração não mudou: o cargo dos DJs já era `%s`",
			role.Name,
		)
	}
}
