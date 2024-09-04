package musiccmds

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/music"
	"github.com/zanz1n/duvua-bot/internal/utils"
	"github.com/zanz1n/duvua-bot/pkg/player"
)

var queueCommandData = discordgo.ApplicationCommand{
	Name:        "queue",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Comandos relacionados à fila",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Commands related to the queue",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "Exibe todas as músicas que estão na fila",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Shows all the musics that are in the queue",
			},
		},
	},
}

func NewQueueCommand(r music.MusicConfigRepository, client *player.HttpClient) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &queueCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &QueueCommand{r: r, c: client},
	}
}

type QueueCommand struct {
	r music.MusicConfigRepository
	c *player.HttpClient
}

func (c *QueueCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		ids := strings.Split(data.CustomID, "/")
		if 2 > len(ids) {
			return errors.New("interação inválida")
		}

		switch ids[1] {
		case "list":
			return c.handleList(s, i)

		case "remove":
			if len(ids) != 3 {
				return errors.New("interação inválida")
			}

			uid := ids[2]
			return c.handleRemove(s, i, uid)

		default:
			return errors.New("interação inválida")
		}
	}

	subCommand, err := i.GetSubCommand()
	if err != nil {
		return err
	}

	switch subCommand.Name {
	case "list":
		return c.handleList(s, i)

	default:
		return errors.New("opção `sub-command` inválida")
	}
}

func (c *QueueCommand) handleList(
	s *discordgo.Session,
	i *manager.InteractionCreate,
) error {
	tracks, err := c.c.GetTracks(i.GuildID)
	if err != nil {
		return err
	}

	fields := make([]*discordgo.MessageEmbedField, len(tracks))

	totalDuration := time.Duration(0)
	for i, track := range tracks {
		value := fmt.Sprintf("**[%s](%s)**", track.Data.Name, track.Data.URL)
		progress := track.State.Progress.Load()

		if track.State != nil {
			totalDuration += track.Data.Duration - progress

			status := "Tocando"
			if track.State.Looping {
				status = "Em loop"
			}
			fields[i] = &discordgo.MessageEmbedField{
				Name: fmt.Sprintf(
					"[%s] Progresso: [%s/%s]",
					status,
					utils.FmtDuration(progress),
					utils.FmtDuration(track.Data.Duration),
				),
				Value: value,
			}
		} else {
			totalDuration += track.Data.Duration

			fields[i] = &discordgo.MessageEmbedField{
				Name: fmt.Sprintf(
					"[%d°] Duração: [%s]",
					i+1,
					utils.FmtDuration(track.Data.Duration),
				),
				Value: value,
			}
		}
	}

	return i.Reply(s, &manager.InteractionResponse{
		Embeds: []*discordgo.MessageEmbed{{
			Title: "Fila de músicas",
			Description: fmt.Sprintf(
				"Duração total da playlist: **[%s]**",
				utils.FmtDuration(totalDuration),
			),
			Fields: fields,
		}},
	})
}

func (c *QueueCommand) handleRemove(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	uid string,
) error {
	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if err = canControl(i.Member, cfg); err != nil {
		return err
	}

	id, err := uuid.Parse(uid)
	if err != nil {
		return errors.New("id da música inválido")
	}

	track, err := c.c.RemoveTrack(i.GuildID, id)
	if err != nil {
		return err
	}

	return i.Replyf(s,
		"Música **[%s](%s)** removida da fila",
		track.Data.Name,
		track.Data.URL,
	)
}
