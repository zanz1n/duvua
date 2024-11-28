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

var unpauseCommandData = discordgo.ApplicationCommand{
	Name:        "unpause",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Despausa a música que está tocando",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Unpauses the music that is playing",
	},
}

func NewUnpauseCommand(r music.MusicConfigRepository, client player.PlayerClient) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &unpauseCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &UnpauseCommand{r: r, c: client},
	}
}

type UnpauseCommand struct {
	r music.MusicConfigRepository
	c player.PlayerClient
}

func (c *UnpauseCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
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

	changed, err := c.c.Unpause(ctx, &player.GuildIdRequest{
		GuildId: cuint64(i.GuildID),
	})
	if err != nil {
		return err
	}

	msg := "Fila"
	if !changed.Changed {
		msg += " já estava"
	}
	msg += " despausada!"

	return i.Replyf(s, msg)
}
