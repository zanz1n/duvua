package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/manager"
)

var helpCommandData = discordgo.ApplicationCommand{
	Name:        "help",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Exibe a lista de todos os comandos do bot e suas funções",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Shows the list of all commands and their functions",
	},
}

func NewHelpCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:       &helpCommandData,
		Category:   manager.CommandCategoryInfo,
		NeedsDefer: false,
		Handler:    &HelpCommand{},
	}
}

type HelpCommand struct {
	m *manager.Manager
}

func (c *HelpCommand) renderHome(i *discordgo.InteractionCreate) discordgo.MessageEmbed {
	var avatarUrl string
	if i.User == nil {
		avatarUrl = i.Member.AvatarURL("128")
	} else {
		avatarUrl = i.User.AvatarURL("128")
	}

	var userName string
	if i.User == nil {
		userName = i.Member.DisplayName()
	} else {
		userName = i.User.GlobalName
	}

	return discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeArticle,
		Title:       "Help",
		Description: "Use o seletor a baixo para ver os comandos disponíveis em cada categoria",
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Requisitado por " + userName,
			IconURL: avatarUrl,
		},
	}
}

func (c *HelpCommand) renderCategory(cat manager.CommandCategory) discordgo.MessageEmbed {
	var catName string

	switch cat {
	case manager.CommandCategoryInfo:
		catName = "Info"
	case manager.CommandCategoryConfig:
		catName = "Config"
	case manager.CommandCategoryFun:
		catName = "Fun"
	case manager.CommandCategoryTicket:
		catName = "Info"
	}

	embed := discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeArticle,
		Title: "Help | " + catName,
	}

	cmds := c.m.GetDataByCategory(manager.CommandAccept{Slash: true, Button: false}, cat)

	fields := make([]*discordgo.MessageEmbedField, len(cmds))

	for i, cmd := range cmds {
		fields[i] = &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%v. %s", i+1, cmd.Name),
			Value:  cmd.Description,
			Inline: true,
		}
	}

	embed.Fields = fields

	return embed
}

func (c *HelpCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		if data.Values == nil || len(data.Values) != 1 {
			return nil
		}

		value := data.Values[0]

		var embed discordgo.MessageEmbed

		switch value {
		case "help":
			embed = c.renderHome(i)
		case "info":
			embed = c.renderCategory(manager.CommandCategoryInfo)
		case "config":
			embed = c.renderCategory(manager.CommandCategoryConfig)
		case "fun":
			embed = c.renderCategory(manager.CommandCategoryFun)
		case "ticket":
			embed = c.renderCategory(manager.CommandCategoryTicket)
		default:
			embed = c.renderHome(i)
		}

		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      i.Message.ID,
			Channel: i.Message.ChannelID,
			Embeds:  &[]*discordgo.MessageEmbed{&embed},
		})
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	} else if i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		options := []discordgo.SelectMenuOption{
			{Label: "Help", Value: "help", Description: "Informações sobre o bot", Emoji: emoji("🏡")},
			{Label: "Info", Value: "info", Description: "Comandos de informação", Emoji: emoji("ℹ️")},
			{Label: "Config", Value: "config", Description: "Comandos de configuração", Emoji: emoji("⚙️")},
			{Label: "Fun", Value: "fun", Description: "Comandos para descontrair", Emoji: emoji("🎉")},
			{Label: "Ticket", Value: "ticket", Description: "Comandos de ticket", Emoji: emoji("🎫")},
		}

		embed := c.renderHome(i)

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    "help",
								Placeholder: "Selecione uma categoria!",
								MenuType:    discordgo.StringSelectMenu,
								MaxValues:   1,
								Options:     options,
							},
						},
					},
				},
				Embeds: []*discordgo.MessageEmbed{&embed},
			},
		})
	}

	return nil
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
