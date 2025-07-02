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

var pauseCommandData = discordgo.ApplicationCommand{
	Name:        "pause",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Pausa a música que está tocando",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Pauses the music that is playing",
	},
}

func NewPauseCommand(r music.MusicConfigRepository, client player.PlayerClient) *manager.Command {
	return &manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &pauseCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &PauseCommand{r: r, c: client},
	}
}

type PauseCommand struct {
	r music.MusicConfigRepository
	c player.PlayerClient
}

func (c *PauseCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
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

	changed, err := c.c.Pause(ctx, &player.GuildIdRequest{
		GuildId: cuint64(i.GuildID),
	})
	if err != nil {
		return err
	}

	msg := "Fila"
	if !changed.Changed {
		msg += " já estava"
	}
	msg += " pausada!"

	return i.Reply(s, &manager.InteractionResponse{
		Content: msg,
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label:    "Despausar",
				Emoji:    emoji("▶️"),
				Style:    discordgo.PrimaryButton,
				CustomID: "unpause",
			}},
		}},
	})
}
