package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

var avatarCommandData = discordgo.ApplicationCommand{
	Name:        "avatar",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Exibe o avatar de um usuário ou o seu",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Displays some user's avatar or yours",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "O usuário cujo avatar deseja ver",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Whose avatar you want to show",
			},
		},
	},
}

func NewAvatarCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:       &avatarCommandData,
		Category:   manager.CommandCategoryFun,
		NeedsDefer: false,
		Handler:    &AvatarCommand{},
	}
}

type AvatarCommand struct {
}

func (c *AvatarCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand &&
		i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return errors.New("interação de tipo inesperado")
	}

	data := i.ApplicationCommandData()

	var (
		name      string
		avatarUrl string
	)
	if opt := utils.GetOption(data.Options, "user"); opt != nil {
		if opt.Type != discordgo.ApplicationCommandOptionUser {
			return errors.New("opção `user` precisa ser um usuário válido")
		}

		if i.Member != nil {
			member, err := s.GuildMember(i.GuildID, opt.UserValue(nil).ID)
			if err != nil {
				return err
			}

			name = member.DisplayName()
			avatarUrl = member.AvatarURL("256")
		} else {
			user := opt.UserValue(s)
			name = user.GlobalName
			avatarUrl = user.AvatarURL("256")
		}
	} else {
		name = utils.CallerNameFromInteraction(i.Interaction)
		avatarUrl = utils.AvatarUrlFromInteraction(i.Interaction, "256")
	}

	embed := discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeArticle,
		Title: "Avatar de " + name,
		Image: &discordgo.MessageEmbedImage{
			URL: avatarUrl,
		},
		Footer: utils.EmbedRequestedByFooter(i.Interaction),
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{&embed},
		},
	})
}
