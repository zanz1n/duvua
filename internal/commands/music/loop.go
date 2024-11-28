package musiccmds

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/pkg/pb/player"
)

var loopCommandData = discordgo.ApplicationCommand{
	Name:        "loop",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Habilita ou desabilita o loop na playlist",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Enables or disables loop in the playlist",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "on",
			Description: "Habilita o loop na playlist",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Enables loop in the playlist",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "off",
			Description: "Desabilita o loop na playlist",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Disables loop in the playlist",
			},
		},
	},
}

func NewLoopCommand(r music.MusicConfigRepository, client player.PlayerClient) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &loopCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &LoopCommand{r: r, c: client},
	}
}

type LoopCommand struct {
	r music.MusicConfigRepository
	c player.PlayerClient
}

func (c *LoopCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	enable := false
	if i.Type == discordgo.InteractionApplicationCommand {
		enable = i.ApplicationCommandData().Name == "on"
	} else if i.Type == discordgo.InteractionMessageComponent {
		enable = i.MessageComponentData().CustomID == "loop/on"
	} else {
		return errors.New("interação inválida")
	}

	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if err = canControl(i.Member, cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	changed, err := c.c.EnableLoop(ctx, &player.EnableLoopRequest{
		GuildId: cuint64(i.GuildID),
		Enable:  enable,
	})
	if err != nil {
		return err
	}

	msg := "Loop"
	if !changed.Changed {
		msg += " já estava"
	}
	if enable {
		msg += " habilitado"
	} else {
		msg += " desabilitado"
	}

	res := manager.InteractionResponse{
		Content: msg,
	}

	if enable {
		res.Components = append(res.Components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label:    "Desabilitar",
				Emoji:    emoji("🔁"),
				Style:    discordgo.PrimaryButton,
				CustomID: "loop/off",
			}},
		})
	}

	return i.Reply(s, &res)
}
