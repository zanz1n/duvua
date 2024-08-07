package commands

import (
	"crypto/sha1"
	"encoding/binary"
	"log/slog"
	"math"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

var shipCommandData = discordgo.ApplicationCommand{
	Name:        "ship",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Mostra a porcentagem de um casal dar certo",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Shows the percentage of a couple working out",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user1",
			Description: "O primeiro usuário",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The first user",
			},
			Required: true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user2",
			Description: "O segundo usuário",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The second user",
			},
			Required: true,
		},
	},
}

func NewShipCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:       &shipCommandData,
		Category:   manager.CommandCategoryFun,
		NeedsDefer: false,
		Handler:    &ShipCommand{},
	}
}

type ShipCommand struct {
}

func (c *ShipCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand &&
		i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return errors.New("interação de tipo inesperado")
	}

	data := i.ApplicationCommandData()

	user1Opt := utils.GetOption(data.Options, "user1")
	if user1Opt == nil {
		return errors.New("opção `user1` é necessária")
	} else if user1Opt.Type != discordgo.ApplicationCommandOptionUser {
		return errors.New("opção `user1` precisa ser um usuário válido")
	}
	user1 := user1Opt.UserValue(nil)

	user2Opt := utils.GetOption(data.Options, "user2")
	if user2Opt == nil {
		return errors.New("opção `user2` é necessária")
	} else if user2Opt.Type != discordgo.ApplicationCommandOptionUser {
		return errors.New("opção `user2` precisa ser um usuário válido")
	}
	user2 := user2Opt.UserValue(nil)

	if user1.ID == s.State.User.ID || user2.ID == s.State.User.ID {
		return s.InteractionRespond(i.Interaction, utils.BasicResponse("Sai pra lá!"))
	} else if user1.ID == user2.ID {
		return s.InteractionRespond(i.Interaction, utils.BasicResponse(
			"<@%v> já faz isso no banheiro todo dia!", user1.ID,
		))
	}

	user1Id, err := strconv.ParseUint(user1.ID, 10, 0)
	if err != nil {
		slog.Error("Failed to parse discord snowflake", "value", user1.ID, "error", err)
		return errors.New("opção `user1` precisa ser um usuário válido")
	}
	user2Id, err := strconv.ParseUint(user2.ID, 10, 0)
	if err != nil {
		slog.Error("Failed to parse discord snowflake", "value", user2.ID, "error", err)
		return errors.New("opção `user2` precisa ser um usuário válido")
	}

	percentage := c.shipPercentage(user1Id, user2Id)

	return s.InteractionRespond(i.Interaction, utils.BasicResponse(
		"<@%s> e <@%s> possuem um chance de %v%s de dar certo",
		user1.ID, user2.ID, percentage, "%",
	))
}

func (c *ShipCommand) shipPercentage(user1 uint64, user2 uint64) int8 {
	sha := sha1.New()
	buf := make([]byte, 0, 16)

	switch {
	case user2 > user1:
		buf = binary.BigEndian.AppendUint64(buf, user1)
		buf = binary.BigEndian.AppendUint64(buf, user2)
	case user1 > user2:
		buf = binary.BigEndian.AppendUint64(buf, user2)
		buf = binary.BigEndian.AppendUint64(buf, user1)
	default:
		return -1
	}

	_, err := sha.Write(buf)
	if err != nil {
		panic(err)
	}

	shasum := sha.Sum(nil)
	shasum = shasum[:8]
	shasum[0] = 0
	shasum[1] = 0

	num := float64(binary.BigEndian.Uint64(shasum))

	rounded := math.Round((num / float64((1 << 48))) * 100)
	percentage := int8(rounded)
	if percentage > 100 {
		slog.Error(
			"Ship generated value above 100%",
			"value_f64", rounded,
			"value_u8", percentage,
		)
		percentage = 100
	} else if 0 > percentage {
		slog.Error(
			"Ship generated value less than 0%",
			"value_f64", rounded,
			"value_u8", percentage,
		)
		percentage = 0
	}

	return percentage
}
