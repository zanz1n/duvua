package musiccmds

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/player"
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

func NewPlayCommand(r music.MusicConfigRepository, client *player.HttpClient) manager.Command {
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
	c *player.HttpClient
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

	tracksData, err := c.c.FetchTrack(query)
	if err != nil {
		return err
	}

	tracks, err := c.c.AddTrack(i.GuildID, player.AddTrackData{
		UserId:        i.Member.User.ID,
		ChannelId:     vs.ChannelID,
		TextChannelId: i.ChannelID,
		Data:          tracksData,
	})
	if err != nil {
		return err
	}

	if len(tracks) == 1 {
		track := tracks[0]

		msg := fmt.Sprintf("Música **[%s](%s) [%s]** adicionada à fila",
			track.Data.Name, track.Data.URL, utils.FmtDuration(track.Data.Duration),
		)

		return i.Reply(s, &manager.InteractionResponse{
			Content: msg,
			Components: []discordgo.MessageComponent{discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Remover",
						Emoji:    emoji("✖️"),
						Style:    discordgo.DangerButton,
						CustomID: "queue/remove/" + track.ID.String(),
					},
				},
			}},
		})
	}

	return i.Replyf(s, "%d músicas adicionadas à fila", len(tracks))
}
