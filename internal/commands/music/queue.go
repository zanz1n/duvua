package musiccmds

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/pb/player"
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
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove",
			Description: "Remove uma música da fila com base na posição",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Removes a music from the queue by its position",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "position",
					Description: "A posição da música que deseja remover",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "The position of the music you want to remove",
					},
					Required: true,
				},
			},
		},
	},
}

func NewQueueCommand(r music.MusicConfigRepository, client player.PlayerClient) manager.Command {
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
	c player.PlayerClient
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

	case "remove":
		pos, err := i.GetIntegerOption("position", true)
		if err != nil {
			return err
		}

		return c.handleRemoveByPosition(s, i, int(pos))

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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := c.c.GetAll(ctx, &player.GetAllRequest{
		GuildId: cuint64(guildId),
		Offset:  int32(paddedOff),
		Limit:   pageSize,
	})
	if err != nil {
		return nil, nil, err
	}

	if offset > int(data.TotalSize) {
		return nil, nil, errors.New("interação inválida")
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(data.Tracks)+1)
	totalDuration := time.Duration(0)

	if page == 0 && data.Playing != nil {
		progress := data.Playing.State.Progress.AsDuration()

		totalDuration += data.Playing.Data.Duration.AsDuration() - progress

		status := "Tocando"
		if data.Playing.State.Looping {
			status = "Em loop"
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("[%s] Progresso: [%s/%s]",
				status,
				utils.FmtDuration(progress),
				utils.FmtDuration(data.Playing.Data.Duration.AsDuration()),
			),
			Value: fmt.Sprintf("**[%s](%s)**",
				data.Playing.Data.Name,
				data.Playing.Data.Url,
			),
		})
	}

	for i, track := range data.Tracks {
		dur := track.Data.Duration.AsDuration()

		value := fmt.Sprintf("**[%s](%s)**", track.Data.Name, track.Data.Url)

		totalDuration += dur

		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf(
				"[%d°] Duração: [%s]",
				offset+i+1,
				utils.FmtDuration(dur),
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
					Disabled: pageSize*(page+1) >= int(data.TotalSize),
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
	id string,
) error {
	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if err = canControl(i.Member, cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	track, err := c.c.Remove(ctx, &player.TrackIdRequest{
		GuildId: cuint64(i.GuildID),
		Id:      id,
	})
	if err != nil {
		return err
	}

	return i.Replyf(s,
		"Música **[%s](<%s>)** removida da fila",
		track.Track.Data.Name,
		track.Track.Data.Url,
	)
}

func (c *QueueCommand) handleRemoveByPosition(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	pos int,
) error {
	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if err = canControl(i.Member, cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	track, err := c.c.RemoveByPosition(ctx, &player.RemoveByPositionRequest{
		GuildId:  cuint64(i.GuildID),
		Position: int32(pos),
	})
	if err != nil {
		return err
	}

	return i.Replyf(s,
		"Música **[%s](<%s>)** removida da fila",
		track.Track.Data.Name,
		track.Track.Data.Url,
	)
}
