package player

import (
	"io"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/pkg/player"
)

type PlayerManager struct {
	players map[uint64]*GuildPlayer
	mu      sync.RWMutex

	s *discordgo.Session
	f *TrackFetcher
	m *PlayerMessenger
}

func NewPlayerManager(s *discordgo.Session, f *TrackFetcher) *PlayerManager {
	return &PlayerManager{
		players: map[uint64]*GuildPlayer{},
		mu:      sync.RWMutex{},
		s:       s,
		f:       f,
		m:       &PlayerMessenger{s: s},
	}
}

func (m *PlayerManager) Get(id uint64) (*GuildPlayer, bool) {
	m.mu.RLock()
	p, ok := m.players[id]
	m.mu.RUnlock()
	return p, ok
}

func (m *PlayerManager) GetOrCreate(id, channelId uint64) *GuildPlayer {
	m.mu.RLock()
	p, ok := m.players[id]
	m.mu.RUnlock()

	if !ok {
		p = newGuildPlayer(id)
		m.mu.Lock()
		m.players[id] = p
		m.mu.Unlock()

		slog.Info(
			"Created guild player",
			"guild_id", p.GuildId,
			"channel_id", channelId,
		)

		go m.guildJobLaunch(p, channelId)
	}

	return p
}

func (m *PlayerManager) Remove(id uint64) {
	m.mu.Lock()
	delete(m.players, id)
	m.mu.Unlock()
}

func (m *PlayerManager) RemoveCheck(id uint64) bool {
	_, ok := m.Get(id)
	if ok {
		m.Remove(id)
	}

	return ok
}

func (m *PlayerManager) guildJobLaunch(p *GuildPlayer, channelId uint64) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(
				"Panic catched while handling queue",
				"guild_id", p.GuildId,
				"error", err,
			)
		}
	}()

	defer func() {
		m.Remove(p.GuildId)
		if p.Interrupt != nil {
			close(p.Interrupt)
		}
	}()

	err := m.guildJob(p, channelId)
	if err != nil {
		slog.Error(
			"Error while handling queue",
			"guild_id", p.GuildId,
			"error", err,
		)
	}
}

func (m *PlayerManager) guildJob(p *GuildPlayer, cId uint64) error {
	const MaxPoolTries = 10
	const PoolTryDelay = time.Second

	start := time.Now()

	guildId := strconv.FormatUint(p.GuildId, 10)
	channelId := strconv.FormatUint(cId, 10)

	vc, err := m.s.ChannelVoiceJoin(guildId, channelId, false, true)
	if err != nil {
		return errors.Unexpectedf(
			"failed to join voice channel `%s`: %s",
			channelId, err,
		)
	}
	defer vc.Disconnect()

	slog.Info("Started queue", "guild_id", guildId, "channel_id", cId)

	defer m.m.OnQueueEnd(p)

	poolTries := 0
	pausedTime := time.Duration(0)
	track := (*player.Track)(nil)

LOOP:
	for {
		if !p.IsLooping() {
			if track = p.Pool(); track == nil {
				poolTries++
				if poolTries >= MaxPoolTries {
					break
				}
				select {
				case <-time.NewTimer(PoolTryDelay).C:
				case evt := <-p.Interrupt:
					if evt == InterruptStop {
						break LOOP
					}
				}
				continue
			} else {
				poolTries = 0
			}
		} else if track == nil {
			break
		}

		slog.Info(
			"Queue started track",
			"track_id", track.ID,
			"guild_id", guildId,
			"queue_size", p.Size(),
		)

		m.m.OnTrackStart(p, track)

		stream, err := m.f.Fetch(track.Data.PlayQuery)
		if err != nil {
			slog.Error("Failed to fetch track", "error", err)
			m.m.OnTrackFailed(p, track)
			continue
		}

		interrupt, pt, err := m.playTrack(vc, p, track, stream)
		if err != nil {
			if err == ErrTooMuchTimePaused {
				break
			} else if err == ErrVoiceConnectionClosed {
				slog.Info("Queue voice connection closed", "guild_id", guildId)
				break
			} else {
				slog.Error(
					"Error while playing track",
					"guild_id", guildId,
					"error", err,
				)
				m.m.OnTrackFailed(p, track)
			}
		}
		pausedTime += pt

		if interrupt == InterruptStop {
			break
		}
	}

	slog.Info(
		"Stopped queue",
		"guild_id", guildId,
		"active_time", (time.Since(start) - pausedTime).Round(time.Millisecond),
		"paused_time", pausedTime.Round(time.Millisecond),
	)

	time.Sleep(500 * time.Millisecond)

	return nil
}

func (m *PlayerManager) playTrack(
	vc *discordgo.VoiceConnection,
	p *GuildPlayer,
	track *player.Track,
	stream Streamer,
) (InterruptType, time.Duration, error) {
	const MaxPausedTime = 5 * 60 * time.Second

	defer stream.Close()

	pausedTime := time.Duration(0)
	timeout := time.NewTicker(time.Second)
	defer timeout.Stop()

	var readStart time.Time
	for {
		readStart = time.Now()
		packet, err := stream.ReadOpus()
		if err != nil {
			if err == io.EOF {
				return InterruptNone, pausedTime, nil
			}
			return InterruptNone, pausedTime, errors.Unexpected(
				"failed to read opus stream: " + err.Error(),
			)
		}

		select {
		case vc.OpusSend <- packet:
			timeout.Reset(time.Second)
			track.State.Progress.Add(time.Since(readStart))

		case evt := <-p.Interrupt:
			timeout.Reset(time.Second)
			if evt != InterruptPause {
				return evt, pausedTime, nil
			}

			pauseStart := time.Now()
			select {
			case e := <-p.Interrupt:
				pausedTime += time.Since(pauseStart)
				if e != InterruptUnpause {
					return e, pausedTime, nil
				}

			case <-time.NewTimer(MaxPausedTime).C:
				pausedTime += time.Since(pauseStart)
				return InterruptNone, pausedTime, ErrTooMuchTimePaused
			}

		case <-timeout.C:
			return InterruptNone, pausedTime, ErrVoiceConnectionClosed
		}
	}
}
