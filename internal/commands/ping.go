package commands

import (
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
)

var pingCommandData = discordgo.ApplicationCommand{
	Name:        "ping",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Responde com pong e mostra o ping do bot",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Replies with pong and shows the bot latency",
	},
}

func NewPingCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &pingCommandData,
		Category: manager.CommandCategoryInfo,
		Handler: &PingCommand{
			total:  atomic.Int32{},
			amount: atomic.Int32{},
		},
	}
}

type PingCommand struct {
	total  atomic.Int32
	amount atomic.Int32
}

func (c *PingCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	var apiLatency int32
	if am := c.amount.Load(); am >= 10 {
		apiLatency = c.total.Load() / am
	} else {
		start := time.Now()

		_, err := s.Client.Get(discordgo.EndpointAPI)
		if err != nil {
			return errors.Unexpected("failed to fetch discord api: " + err.Error())
		}
		took := time.Since(start).Milliseconds()

		apiLatency = c.total.Add(int32(took)) / c.amount.Add(1)
	}

	return i.Replyf(s,
		"🏓 **Pong!**\n🛜 Ping do websocket: %vms\n📡 Ping da api: %vms",
		s.HeartbeatLatency().Milliseconds(),
		apiLatency,
	)
}
