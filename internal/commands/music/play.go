package musiccmds

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/pb/player"
)

var playCommandData = discordgo.ApplicationCommand{
	Name:        "play",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Toca uma música",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Plays a music from internet",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "query",
			Description: "O nome ou a url da música",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The name or the url of the music",
			},
			Required: true,
		},
	},
}

func NewPlayCommand(r music.MusicConfigRepository, client player.PlayerClient) manager.Command {
	if client == nil {
		panic("NewPlayCommand() client must not be nil")
	}

	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &playCommandData,
		Category: manager.CommandCategoryMusic,
		Handler:  &PlayCommand{r: r, c: client},
	}
}

type PlayCommand struct {
	r music.MusicConfigRepository
	c player.PlayerClient
}

func (c *PlayCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	query, err := i.GetStringOption("query", true)
	if err != nil {
		return err
	}

	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if !canPlay(i.Member, cfg) {
		return errors.New("você não tem permissão para tocar músicas no servidor")
	}

	vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return errors.New(
				"você precisa estar em um canal de voz para usar esse comando",
			)
		}
		return err
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1()

	tracksData, err := c.c.Fetch(ctx1, &player.FetchRequest{
		Query: query,
	})
	if err != nil {
		return err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	tracksRes, err := c.c.Add(ctx2, &player.AddRequest{
		GuildId:       cuint64(i.GuildID),
		UserId:        cuint64(i.Member.User.ID),
		ChannelId:     cuint64(vs.ChannelID),
		TextChannelId: cuint64(i.ChannelID),
		Data:          tracksData.Data,
	})
	if err != nil {
		return err
	}

	tracks := tracksRes.Tracks

	if len(tracks) == 1 {
		track := tracks[0]

		msg := fmt.Sprintf("Música **[%s](<%s>) [%s]** adicionada à fila",
			track.Data.Name,
			track.Data.Url,
			utils.FmtDuration(track.Data.Duration.AsDuration()),
		)

		return i.Reply(s, &manager.InteractionResponse{
			Content: msg,
			Components: []discordgo.MessageComponent{discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Remover",
						Emoji:    emoji("✖️"),
						Style:    discordgo.DangerButton,
						CustomID: "queue/remove/" + track.Id,
					},
				},
			}},
		})
	}

	return i.Reply(s, &manager.InteractionResponse{
		Content: fmt.Sprintf("%d músicas adicionadas à fila", len(tracks)),
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Ver fila",
					Emoji:    emoji("📜"),
					Style:    discordgo.PrimaryButton,
					CustomID: "queue/list",
				},
			},
		}},
	})
}
