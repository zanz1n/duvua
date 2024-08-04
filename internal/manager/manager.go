package manager

import (
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/utils"
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

func (m *Manager) handleCommand(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	startTime time.Time,
	cmd *Command,
) {
	if err := cmd.Handler.Handle(s, i); err != nil {
		oerr := err

		expected := false
		if err, ok := err.(errors.Expected); ok {
			expected = err.IsExpected()
		}

		var errorRes string
		if expected {
			errorRes = "Erro: " + err.Error()
		} else {
			errorRes = "Algo deu errado!"
		}

		if cmd.NeedsDefer {
			_, err = s.InteractionResponseEdit(i.Interaction, utils.BasicResponseEdit(errorRes))
		} else {
			err = s.InteractionRespond(i.Interaction, utils.BasicEphemeralResponse(errorRes))
		}
		if err != nil {
			slog.Error(
				"Failed to set command response after error",
				"error", err,
			)
		}

		if !expected {
			slog.Error(
				"Exception caught while executing a command",
				"name", cmd.Data.Name,
				"took", time.Since(startTime),
				"error", oerr,
			)
			return
		}
	}

	slog.Info(
		"Command executed",
		"name", cmd.Data.Name,
		"took", time.Since(startTime),
	)
}

func (m *Manager) handleButton(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	startTime time.Time,
) {
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
}

func (m *Manager) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	startTime := time.Now()

	var (
		name string
		cmd  Command
		ok   bool = false
		btnH bool = false
	)
	if i.Type == discordgo.InteractionApplicationCommand ||
		i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		name = i.ApplicationCommandData().Name

		if cmd, ok = m.cmds[name]; ok {
			if !cmd.Accepts.Slash {
				return
			}
		} else {
			return
		}
	} else if i.Type == discordgo.InteractionMessageComponent {
		name = i.MessageComponentData().CustomID

		if strings.Contains(name, "/") {
			splited := strings.Split(name, "/")
			if len(splited) >= 1 {
				name = splited[0]
			}
		}

		if cmd, ok = m.cmds[name]; ok {
			if !cmd.Accepts.Button {
				btnH = true
			}
		} else {
			btnH = true
		}
	}

	if btnH {
		m.handleButton(s, i, startTime)
		return
	}
	if !ok {
		slog.Info(
			"Cound not find a handler for the given command",
			"name", name,
			"took", time.Since(startTime),
		)
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

	m.handleCommand(s, i, startTime, &cmd)
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
	start := time.Now()

	arr := m.GetData(CommandAccept{Slash: true, Button: false})

	gId := ""
	if guildId != nil {
		gId = *guildId
	}

	created, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, gId, arr)
	if err != nil {
		slog.Error(
			"Something went wrong while posting commands",
			"took", time.Since(start),
			"error", err,
		)

		return
	}

	slog.Info(
		"Posted commands",
		"success_count", len(arr),
		"failed_count", len(arr)-len(created),
		"took", time.Since(start),
	)
}
