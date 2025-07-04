package commands

import (
	configcmds "github.com/zanz1n/duvua/commands/config"
	funcmds "github.com/zanz1n/duvua/commands/fun"
	infocmds "github.com/zanz1n/duvua/commands/info"
	modcmds "github.com/zanz1n/duvua/commands/moderation"
	musiccmds "github.com/zanz1n/duvua/commands/music"
	ticketcmds "github.com/zanz1n/duvua/commands/ticket"
	"github.com/zanz1n/duvua/events"
	"github.com/zanz1n/duvua/internal/anime"
	"github.com/zanz1n/duvua/internal/lang"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/ticket"
	"github.com/zanz1n/duvua/internal/welcome"
	"github.com/zanz1n/duvua/pkg/pb/player"
)

func Wire(
	m *manager.Manager,
	welcomeRepo welcome.WelcomeRepository,
	welcomeEvt *events.MemberAddEvent,
	animeApi *anime.AnimeApi,
	translator lang.Translator,
	ticketRepository ticket.TicketRepository,
	ticketConfigRepository ticket.TicketConfigRepository,
	musicRepository music.MusicConfigRepository,
	musicClient player.PlayerClient,
) {
	m.Add(configcmds.NewWelcomeCommand(welcomeRepo, welcomeEvt))

	m.Add(funcmds.NewAvatarCommand())
	m.Add(funcmds.NewCloneCommand())
	m.Add(funcmds.NewShipCommand())

	m.Add(infocmds.NewFactsCommand())
	m.Add(infocmds.NewHelpCommand(m))
	m.Add(infocmds.NewPingCommand())
	m.Add(infocmds.NewAnimeCommand(animeApi, translator))

	m.Add(modcmds.NewClearCommand())

	m.Add(ticketcmds.NewTicketAdminCommand(ticketRepository, ticketConfigRepository))
	m.Add(ticketcmds.NewTicketCommand(ticketRepository, ticketConfigRepository))

	m.Add(musiccmds.NewMusicAdminCommand(musicRepository))
	m.Add(musiccmds.NewPlayCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewSkipCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewStopCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewQueueCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewLoopCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewPauseCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewUnpauseCommand(musicRepository, musicClient))
}
