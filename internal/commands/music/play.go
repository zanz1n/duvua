package musiccmds

import (
	"fmt"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/music"
	"github.com/zanz1n/duvua-bot/internal/utils"
	"github.com/zanz1n/duvua-bot/pkg/player"
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

	cfg, err := c.r.GetOrDefault(i.GuildID)
	if err != nil {
		return err
	}

	if !canPlay(i.Member, cfg) {
		return errors.New("você não ter permissão para tocar músicas no servidor")
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

	query, err := i.GetStringOption("query", true)
	if err != nil {
		return err
	}

	trackData, err := c.c.FetchTrack(query)
	if err != nil {
		return err
	}

	track, err := c.c.AddTrack(i.GuildID, player.AddTrackData{
		UserId:    i.Member.User.ID,
		ChannelId: vs.ChannelID,
		Data:      trackData,
	})
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Música **[%s](%s)** adicionada à fila\n\n**Duração: [%s]**",
		track.Data.Name, track.Data.URL, fmtDuration(track.Data.Duration),
	)

	return i.Reply(s, &manager.InteractionResponse{
		Embeds: []*discordgo.MessageEmbed{{
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: track.Data.Thumbnail,
			},
			Description: msg,
			Footer:      utils.EmbedRequestedByFooter(i.Interaction),
		}},
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Remover",
					Emoji:    emoji("✖️"),
					Style:    discordgo.DangerButton,
					CustomID: "music/cancel/" + track.ID.String(),
				},
			},
		}},
	})
}

func canPlay(m *discordgo.Member, cfg *music.MusicConfig) bool {
	switch cfg.PlayMode {
	case music.MusicPermissionAll:
		return true
	case music.MusicPermissionAdm:
		return utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator)
	case music.MusicPermissionDJ:
		return utils.HasPerm(m.Permissions, discordgo.PermissionAdministrator) ||
			slices.Contains(m.Roles, cfg.DjRole)
	default:
		return false
	}
}

func fmtDuration(d time.Duration) string {
	if 0 > d {
		return "0s"
	}

	hour := d / time.Hour
	minute := (d - (hour * time.Hour)) / time.Minute
	second := (d - (hour * time.Hour) - (minute * time.Minute)) / time.Second

	switch {
	case hour == 0 && minute == 0:
		return fmt.Sprintf("%ds", second)
	case hour == 0:
		return fmt.Sprintf("%dm:%ds", minute, second)
	default:
		return fmt.Sprintf("%dh:%dm:%ds", hour, minute, second)
	}
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
