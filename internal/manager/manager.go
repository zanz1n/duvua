package manager

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Manager struct {
	cmds          map[string]Command
	buttonHandler InteractionHandler
}

func NewManager() *Manager {
	return &Manager{
		cmds:          make(map[string]Command),
		buttonHandler: &DefaultInteractionHandler{},
	}
}

func (m *Manager) Add(cmd Command) {
	m.cmds[cmd.Data.Name] = cmd
}

func (m *Manager) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	startTime := time.Now()

	var (
		cmd  Command
		ok   bool
		btnH bool = false
	)
	if i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		if cmd, ok = m.cmds[i.ApplicationCommandData().Name]; ok {
			if !cmd.Accepts.Slash {
				return
			}
		} else {
			return
		}
	} else if i.Type == discordgo.InteractionMessageComponent {
		if cmd, ok = m.cmds[i.MessageComponentData().CustomID]; ok {
			if !cmd.Accepts.Button {
				btnH = true
			}
		} else {
			btnH = true
		}
	}

	if btnH {
		if err := m.buttonHandler.Handle(s, i); err != nil {
			slog.Error(
				"Exception caught while handling message component",
				"took", time.Since(startTime),
				"error", err,
			)
		} else {
			slog.Info(
				"Message component action handled",
				"took", time.Since(startTime),
			)
		}

		return
	}

	if cmd.NeedsDefer {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			slog.Error("Failed to defer reply interaction", "error", err)
			return
		}
	}

	if err := cmd.Handler.Handle(s, i); err != nil {
		slog.Error(
			"Exception caught while executing a command",
			"name", cmd.Data.Name,
			"took", time.Since(startTime),
			"error", err,
		)
	} else {
		slog.Info(
			"Command executed",
			"name", cmd.Data.Name,
			"took", time.Since(startTime),
		)
	}
}

func (m *Manager) AutoHandle(s *discordgo.Session) {
	s.AddHandler(m.Handle)
}

func (m *Manager) ButtonHandler(h InteractionHandler) {
	m.buttonHandler = h
}

func (m *Manager) GetData(accepts CommandAccept) []*discordgo.ApplicationCommand {
	arr := []*discordgo.ApplicationCommand{}

	for _, cmd := range m.cmds {
		if cmd.Accepts.Button && accepts.Button {
			arr = append(arr, cmd.Data)
		} else if cmd.Accepts.Slash && accepts.Slash {
			arr = append(arr, cmd.Data)
		}
	}

	return arr
}

func (m *Manager) GetDataByCategory(accepts CommandAccept, category CommandCategory) []*discordgo.ApplicationCommand {
	arr := []*discordgo.ApplicationCommand{}

	for _, cmd := range m.cmds {
		if cmd.Category == category {
			if cmd.Accepts.Button && accepts.Button {
				arr = append(arr, cmd.Data)
			} else if cmd.Accepts.Slash && accepts.Slash {
				arr = append(arr, cmd.Data)
			}
		}
	}

	return arr
}

func (m *Manager) PostCommands(s *discordgo.Session, guildId *string) {
	arr := m.GetData(CommandAccept{Slash: true, Button: false})

	gId := ""
	if guildId != nil {
		gId = *guildId
	}

	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, gId, arr)
	if err != nil {
		slog.Error("Something went wrong while posting commands", "error", err)

		return
	}

	slog.Info(
		"Posted commands",
		"success_count", len(arr),
		"failed_count", len(arr)-len(created),
	)
}
