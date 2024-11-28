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

var skipCommandData = discordgo.ApplicationCommand{
	Name:        "skip",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Pula a música atual",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Skips the current music",
	},
}

func NewSkipCommand(r music.MusicConfigRepository, client player.PlayerClient) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &skipCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &SkipCommand{r: r, c: client},
	}
}

type SkipCommand struct {
	r music.MusicConfigRepository
	c player.PlayerClient
}

func (c *SkipCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
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

	track, err := c.c.Skip(ctx, &player.GuildIdRequest{
		GuildId: cuint64(i.GuildID),
	})
	if err != nil {
		return err
	}

	return i.Replyf(s,
		"Música **[%s](<%s>)** pulada",
		track.Track.Data.Name,
		track.Track.Data.Url,
	)
}
