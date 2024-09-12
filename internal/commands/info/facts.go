package infocmds

import (
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
)

var factsNumberOpt = []*discordgo.ApplicationCommandOption{
	{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "number",
		Description: "O número que deseja pesquisar (caso não informado será aleatório)",
		DescriptionLocalizations: map[discordgo.Locale]string{
			discordgo.EnglishUS: "The number you want to search (if not set will be random)",
		},
	},
}

var factsCommandData = discordgo.ApplicationCommand{
	Name:        "facts",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Exibe curiosidades sobre números",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Shows facts about numbers (if not set will be random)",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "year",
			Description: "Exibe curiosidades sobre um ano",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Show facts about some year",
			},
			Options: factsNumberOpt,
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "number",
			Description: "Exibe curiosidades sobre um número qualquer",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Show facts about some arbitrary number",
			},
			Options: factsNumberOpt,
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "math",
			Description: "Exibe curiosidades de matemática sobre um número qualquer",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Show math facts about some arbitrary number",
			},
			Options: factsNumberOpt,
		},
	},
}

func NewFactsCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &factsCommandData,
		Category: manager.CommandCategoryInfo,
		Handler:  &FactsCommand{},
	}
}

type FactsCommand struct {
}

func (c *FactsCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	subCommand, err := i.GetSubCommand()
	if err != nil {
		return err
	}

	number := "random"
	numberOpt, err := i.GetTypedOption("number", false, discordgo.ApplicationCommandOptionInteger)
	if err != nil {
		return err
	} else if numberOpt != nil {
		number = strconv.FormatInt(numberOpt.IntValue(), 10)
	}

	const baseUrl = "http://numbersapi.com/"
	var url string
	switch subCommand.Name {
	case "year":
		url = baseUrl + number + "/" + "year"
	case "math":
		url = baseUrl + number + "/" + "math"
	default:
		url = baseUrl + number + "/" + "trivia"
	}

	res, err := s.Client.Get(url)
	if err != nil {
		return errors.Unexpected("failed to fetch numbers api: " + err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Unexpectedf(
			"failed to fetch numbers api: status code %v",
			res.StatusCode,
		)
	} else if mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); mt != "text/plain" {
		return errors.Unexpectedf(
			"failed to fetch numbers api: unexpected content type %s",
			mt,
		)
	}

	bodyBuf, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Unexpected("failed to fetch numbers api: " + err.Error())
	}
	bodyS := string(bodyBuf)

	return i.Replyf(s, bodyS)
}
