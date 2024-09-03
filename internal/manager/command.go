package manager

import "github.com/bwmarrin/discordgo"

type CommandCategory uint8

const (
	CommandCategoryInfo CommandCategory = iota + 1
	CommandCategoryConfig
	CommandCategoryFun
	CommandCategoryTicket
	CommandCategoryModeration
	CommandCategoryMusic
)

type InteractionHandler interface {
	Handle(s *discordgo.Session, i *InteractionCreate) error
}

type DefaultInteractionHandler struct{}

func (h *DefaultInteractionHandler) Handle(s *discordgo.Session, i *InteractionCreate) error {
	return nil
}

type CommandAccept struct {
	Slash  bool
	Button bool
}

type Command struct {
	Accepts  CommandAccept
	Data     *discordgo.ApplicationCommand
	Category CommandCategory
	Handler  InteractionHandler
}
