package utils

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func EmbedRequestedByFooter(i *discordgo.Interaction) *discordgo.MessageEmbedFooter {
	return &discordgo.MessageEmbedFooter{
		Text:    "Requisitado por " + CallerNameFromInteraction(i),
		IconURL: AvatarUrlFromInteraction(i, "128"),
	}
}

func CallerNameFromInteraction(i *discordgo.Interaction) string {
	if i.User == nil {
		return i.Member.DisplayName()
	} else {
		return i.User.GlobalName
	}
}

func AvatarUrlFromInteraction(i *discordgo.Interaction, size string) string {
	var avatarUrl string
	if i.User == nil {
		avatarUrl = i.Member.AvatarURL(size)
	} else {
		avatarUrl = i.User.AvatarURL(size)
	}

	return avatarUrl
}

func GetSubCommand(
	opts []*discordgo.ApplicationCommandInteractionDataOption,
) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range opts {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			return opt
		}
	}

	return nil
}

func HasPerm(memberPerm int64, target int64) bool {
	return memberPerm&target == target
}

func GetOption(
	opts []*discordgo.ApplicationCommandInteractionDataOption,
	name string,
) *discordgo.ApplicationCommandInteractionDataOption {
	for _, opt := range opts {
		if opt.Name == name {
			return opt
		}
	}

	return nil
}

func BasicResponse(format string, args ...any) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(format, args...),
		},
	}
}

func BasicEphemeralResponse(format string, args ...any) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf(format, args...),
		},
	}
}

func BasicResponseEdit(format string, args ...any) *discordgo.WebhookEdit {
	fmt := fmt.Sprintf(format, args...)
	return &discordgo.WebhookEdit{
		Content: &fmt,
	}
}

type StatusType uint8

const (
	StatusTypeStarting StatusType = 96
	StatusTypeStopping StatusType = 73
	StatusTypeIdle     StatusType = 24
)

func SetStatus(s *discordgo.Session, status StatusType) {
	str, name := "online", "/help"

	if status == StatusTypeIdle {
		str, name = "online", "/help"
	} else if status == StatusTypeStarting {
		str, name = "dnd", "Iniciando ..."
	} else if status == StatusTypeStopping {
		str, name = "dnd", "Desligando ..."
	} else {
		slog.Error("Failed to parse StatusType enumeration. Invalid")
	}

	s.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince: nil,
		Status:    str,
		AFK:       false,
		Activities: []*discordgo.Activity{
			{
				Name: name,
				Type: discordgo.ActivityTypeGame,
			},
		},
	})
}
