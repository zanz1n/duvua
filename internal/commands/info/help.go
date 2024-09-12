package infocmds

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/utils"
)

var helpCommandData = discordgo.ApplicationCommand{
	Name:        "help",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Exibe a lista de todos os comandos do bot e suas funções",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Shows the list of all commands and their functions",
	},
}

func NewHelpCommand(m *manager.Manager) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &helpCommandData,
		Category: manager.CommandCategoryInfo,
		Handler:  &HelpCommand{m: m},
	}
}

type HelpCommand struct {
	m *manager.Manager
}

func (c *HelpCommand) renderHome(i *manager.InteractionCreate) discordgo.MessageEmbed {
	return discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeArticle,
		Title:       "Help",
		Description: "Use o seletor a baixo para ver os comandos disponíveis em cada categoria",
		Footer:      utils.EmbedRequestedByFooter(i.Interaction),
	}
}

func (c *HelpCommand) renderCategory(cat manager.CommandCategory) discordgo.MessageEmbed {
	var catName string

	switch cat {
	case manager.CommandCategoryInfo:
		catName = "Info"
	case manager.CommandCategoryMusic:
		catName = "Música"
	case manager.CommandCategoryConfig:
		catName = "Config"
	case manager.CommandCategoryFun:
		catName = "Fun"
	case manager.CommandCategoryTicket:
		catName = "Ticket"
	case manager.CommandCategoryModeration:
		catName = "Moderação"
	}

	embed := discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeArticle,
		Title: "Help | " + catName,
	}

	cmds := c.m.GetDataByCategory(manager.CommandAccept{Slash: true, Button: false}, cat)

	fields := []*discordgo.MessageEmbedField{}

	for _, cmd := range cmds {
		if len(cmd.Options) > 0 {
			kind := cmd.Options[0].Type
			isRootCmd := kind == discordgo.ApplicationCommandOptionSubCommand ||
				kind == discordgo.ApplicationCommandOptionSubCommandGroup

			if isRootCmd {
				newFields := c.appendSubCommandFields(cmd)
				fields = append(fields, newFields...)
				continue
			}
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("/" + cmd.Name),
			Value:  cmd.Description,
			Inline: true,
		})
	}

	embed.Fields = fields

	return embed
}

func (c *HelpCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		if data.Values == nil || len(data.Values) != 1 {
			return errors.New("o input selecionado é inválido")
		}

		value := data.Values[0]

		var embed discordgo.MessageEmbed

		switch value {
		case "help":
			embed = c.renderHome(i)
		case "info":
			embed = c.renderCategory(manager.CommandCategoryInfo)
		case "music":
			embed = c.renderCategory(manager.CommandCategoryMusic)
		case "config":
			embed = c.renderCategory(manager.CommandCategoryConfig)
		case "fun":
			embed = c.renderCategory(manager.CommandCategoryFun)
		case "ticket":
			embed = c.renderCategory(manager.CommandCategoryTicket)
		case "moderation":
			embed = c.renderCategory(manager.CommandCategoryModeration)
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

		return i.DeferUpdate(s)
	} else if i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		options := []discordgo.SelectMenuOption{
			// {Label: "Help", Value: "help", Description: "Informações sobre o bot", Emoji: emoji("🏡")},
			{Label: "Info", Value: "info", Description: "Comandos de informação", Emoji: emoji("ℹ️")},
			{Label: "Música", Value: "music", Description: "Comandos de música", Emoji: emoji("🎵")},
			{Label: "Config", Value: "config", Description: "Comandos de configuração", Emoji: emoji("⚙️")},
			{Label: "Fun", Value: "fun", Description: "Comandos para descontrair", Emoji: emoji("🎉")},
			{Label: "Moderação", Value: "moderation", Description: "Comandos de moderação", Emoji: emoji("🔨")},
			{Label: "Ticket", Value: "ticket", Description: "Comandos de ticket", Emoji: emoji("🎫")},
		}

		embed := c.renderHome(i)

		return i.Reply(s, &manager.InteractionResponse{
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
		})
	}

	return nil
}

func (c *HelpCommand) appendSubCommandFields(
	cmd *discordgo.ApplicationCommand,
) []*discordgo.MessageEmbedField {
	fields := []*discordgo.MessageEmbedField{}

	for _, opt := range cmd.Options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("/%s %s", cmd.Name, opt.Name),
				Value:  opt.Description,
				Inline: true,
			})
		} else if opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			for _, subopt := range opt.Options {
				if subopt.Type == discordgo.ApplicationCommandOptionSubCommand {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("/%s %s %s", cmd.Name, opt.Name, subopt.Name),
						Value:  subopt.Description,
						Inline: true,
					})
				}
			}
		}
	}

	return fields
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
