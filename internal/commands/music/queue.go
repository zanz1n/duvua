package musiccmds

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/player"
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
			off := 0
			if len(ids) > 2 {
				off, _ = strconv.Atoi(ids[2])
			}

			embeds, components, err := c.handleList(i.GuildID, off)
			if err != nil {
				return err
			}

			if err = i.DeferUpdate(s); err != nil {
				return err
			}

			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:         i.Message.ID,
				Channel:    i.Message.ChannelID,
				Embeds:     &embeds,
				Components: &components,
			})
			return err

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
		embeds, components, err := c.handleList(i.GuildID, 0)
		if err != nil {
			return err
		}

		return i.Reply(s, &manager.InteractionResponse{
			Embeds:     embeds,
			Components: components,
		})

	default:
		return errors.New("opção `sub-command` inválida")
	}
}

func (c *QueueCommand) handleList(
	guildId string,
	page int,
) ([]*discordgo.MessageEmbed, []discordgo.MessageComponent, error) {
	const pageSize = 10

	offset, paddedOff := pageSize*page, pageSize*page

	data, err := c.c.GetTracks(guildId, paddedOff, pageSize)
	if err != nil {
		return nil, nil, err
	}

	if offset > data.TotalSize {
		return nil, nil, errors.New("interação inválida")
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(data.Tracks)+1)
	totalDuration := time.Duration(0)

	if page == 0 && data.Playing != nil {
		progress := data.Playing.State.Progress.Load()

		totalDuration += data.Playing.Data.Duration - progress

		status := "Tocando"
		if data.Playing.State.Looping {
			status = "Em loop"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("[%s] Progresso: [%s/%s]",
				status,
				utils.FmtDuration(progress),
				utils.FmtDuration(data.Playing.Data.Duration),
			),
			Value: fmt.Sprintf("**[%s](%s)**",
				data.Playing.Data.Name,
				data.Playing.Data.URL,
			),
		})
	}

	for i, track := range data.Tracks {
		value := fmt.Sprintf("**[%s](%s)**", track.Data.Name, track.Data.URL)

		totalDuration += track.Data.Duration

		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf(
				"[%d°] Duração: [%s]",
				offset+i+1,
				utils.FmtDuration(track.Data.Duration),
			),
			Value: value,
		})
	}

	title := "Fila de músicas"
	if data.TotalSize > pageSize {
		title += fmt.Sprintf(". Pág. %d/%d", page+1, (data.TotalSize/pageSize)+1)
	}

	embeds := []*discordgo.MessageEmbed{{
		Title: title,
		Description: fmt.Sprintf(
			"Duração total da playlist: **[%s]**",
			utils.FmtDuration(totalDuration),
		),
		Fields: fields,
	}}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Anterior",
					Emoji:    emoji("◀️"),
					Style:    discordgo.PrimaryButton,
					CustomID: "queue/list/" + strconv.Itoa(page-1),
					Disabled: 0 >= page,
				},
				discordgo.Button{
					Label:    "Próximo",
					Emoji:    emoji("▶️"),
					Style:    discordgo.PrimaryButton,
					Disabled: pageSize*(page+1) >= data.TotalSize,
					CustomID: "queue/list/" + strconv.Itoa(page+1),
				},
			},
		},
	}

	return embeds, components, nil
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
