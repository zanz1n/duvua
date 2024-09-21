package musiccmds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/pkg/player"
)

var skipCommandData = discordgo.ApplicationCommand{
	Name:        "skip",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Pula a música atual",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Skips the current music",
	},
}

func NewSkipCommand(r music.MusicConfigRepository, client *player.HttpClient) manager.Command {
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
	c *player.HttpClient
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

	track, err := c.c.Skip(i.GuildID)
	if err != nil {
		return err
	}

	return i.Replyf(s,
		"Música **[%s](<%s>)** pulada",
		track.Data.Name,
		track.Data.URL,
	)
}
