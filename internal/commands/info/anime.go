package infocmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/anime"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

var animeCommandData = discordgo.ApplicationCommand{
	Name:        "anime",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Pesquisa por um anime e exibe informações sobre ele",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Searches for an anime and shows information about it",
	},
	Options: []*discordgo.ApplicationCommandOption{{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "name",
		Description: "O nome do anime",
		DescriptionLocalizations: map[discordgo.Locale]string{
			discordgo.EnglishUS: "The name of the anime",
		},
		Required: true,
	}},
}

func NewAnimeCommand(a *anime.AnimeApi) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &animeCommandData,
		Category: manager.CommandCategoryInfo,
		Handler:  &AnimeCommand{a: a},
	}
}

type AnimeCommand struct {
	a *anime.AnimeApi
}

func (c *AnimeCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Type == discordgo.InteractionMessageComponent {
		return c.handleComponent(s, i)
	}

	name, err := i.GetStringOption("name", true)
	if err != nil {
		return err
	}

	a, err := c.a.GetByName(name)
	if err != nil {
		return err
	}

	embed := discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeArticle,
		Title:  a.Attributes.CanonicalTitle,
		Footer: utils.EmbedRequestedByFooter(i.Interaction),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    a.Attributes.PosterImage.Small,
			Width:  int(a.Attributes.PosterImage.Meta.Dimensions.Small.Width),
			Height: int(a.Attributes.PosterImage.Meta.Dimensions.Small.Height),
		},
		Description: fmt.Sprintf(
			"ID: %d\nTrailer: https://youtu.be/%s",
			a.ID,
			a.Attributes.YoutubeVideoId,
		),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📺 Tipo",
				Value:  a.Attributes.Subtype.StringPtBr(),
				Inline: true,
			},
			{
				Name:   "💾 Status",
				Value:  a.Attributes.Status.StringPtBr(),
				Inline: true,
			},
			{
				Name:   "📆 Começo",
				Value:  a.Attributes.StartDate.StringPtBr(),
				Inline: true,
			},
			{
				Name:   "📆 Término",
				Value:  a.Attributes.EndDate.StringPtBr(),
				Inline: true,
			},
			{
				Name:   "🎬 Episódios",
				Value:  strconv.Itoa(int(a.Attributes.EpisodeCount)),
				Inline: true,
			},
			{
				Name:   "⏲️ Duração dos episódios",
				Value:  strconv.Itoa(int(a.Attributes.EpisodeLength)),
				Inline: true,
			},
			{
				Name:   "👶 Classificação indicativa",
				Value:  a.Attributes.AgeRating.StringPtBr(),
				Inline: true,
			},
			{
				Name:   "🔞 NSFW",
				Value:  strconv.FormatBool(a.Attributes.NSFW),
				Inline: true,
			},
			{
				Name:   "💯  Nota",
				Value:  fmt.Sprintf("%s/100", a.Attributes.AverageRating),
				Inline: true,
			},
			{
				Name:   "⭐ Favoritos",
				Value:  strconv.Itoa(int(a.Attributes.FavoritesCount)),
				Inline: true,
			},
			{
				Name:   "🙎 Usuários",
				Value:  strconv.Itoa(int(a.Attributes.UserCount)),
				Inline: true,
			},
			{
				Name:   "📰 Popularidade",
				Value:  fmt.Sprintf("%d°", a.Attributes.PopularityRank),
				Inline: true,
			},
		},
	}
	var components []discordgo.MessageComponent

	if a.Attributes.Titles.JapanJapanese != "" && a.Attributes.Titles.English != "" {
		embed.Title = a.Attributes.Titles.English +
			" | " + a.Attributes.Titles.JapanJapanese
	}

	if len(a.Attributes.Synopsis) > 1023 {
		a.Attributes.Synopsis = a.Attributes.Synopsis[0:1018] + "[...]"
		components = []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label:    "Ver sinopse completa",
				Emoji:    emoji("📰"),
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("anime/%d", a.ID),
			}},
		}}

	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:  "Sinopse (EN_US)",
		Value: a.Attributes.Synopsis,
	})

	return i.Reply(s, &manager.InteractionResponse{
		Embeds:     []*discordgo.MessageEmbed{&embed},
		Components: components,
	})
}

func (c *AnimeCommand) handleComponent(s *discordgo.Session, i *manager.InteractionCreate) error {
	data := i.MessageComponentData()

	split := strings.Split(data.CustomID, "/")
	if len(split) != 2 {
		return errors.New("interação inválida")
	}

	id, err := strconv.ParseInt(split[1], 10, 0)
	if err != nil {
		return errors.New("interação inválida")
	}

	a, err := c.a.GetById(id)
	if err != nil {
		return err
	}

	return i.Replyf(s, a.Attributes.Synopsis)
}