package utils

import (
	"log/slog"
	"time"

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

func HasPerm(memberPerm int64, target int64) bool {
	return memberPerm&target == target
}

type StatusType uint8

const (
	StatusTypeStarting StatusType = 96
	StatusTypeStopping StatusType = 73
	StatusTypeIdle     StatusType = 24
)

func SetStatus(s *discordgo.Session, status StatusType) {
	start := time.Now()

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

	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
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
	if err != nil {
		slog.Error(
			"Failed to set bot status",
			"took", time.Since(start),
			"error", err,
		)
	} else {
		slog.Info(
			"Bot status set",
			"status", str,
			"name", name,
			"took", time.Since(start),
		)
	}
}
