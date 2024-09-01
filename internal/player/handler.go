package player

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zanz1n/duvua-bot/pkg/player"
)

type Handler struct {
	m *PlayerManager
	f *TrackFetcher
}

func NewHandler(manager *PlayerManager, f *TrackFetcher) *Handler {
	return &Handler{m: manager, f: f}
}

func (h *Handler) FetchTrack(query string) (*player.TrackData, error) {
	return h.f.Search(query)
}

func (h *Handler) AddTrack(guildId uint64, data player.AddTrackData) player.Track {
	track := player.NewTrack(data.UserId, data.ChannelId, data.Data)
	p := h.m.GetOrCreate(guildId, data.ChannelId)
	p.AddTrack(track)

	slog.Info(
		"Added track to queue",
		"guild_id", guildId,
		"user_id", data.UserId,
		"url", track.Data.PlayQuery,
		"duration", track.Data.Duration.Round(time.Millisecond),
	)

	return track
}

func (h *Handler) GetPlayingTrack(guildId uint64) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, ErrNoActivePlayer
	}

	current, ok := p.GetCurrent()
	if !ok {
		return nil, ErrTrackNotFoundInQueue
	}
	return current, nil
}

func (h *Handler) Skip(guildId uint64) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, ErrNoActivePlayer
	}

	t := p.Skip()
	if t == nil {
		return nil, ErrNoActivePlayer
	}

	return t, nil
}

func (h *Handler) Stop(guildId uint64) error {
	p, ok := h.m.Get(guildId)
	if !ok {
		return ErrNoActivePlayer
	}

	p.Stop()
	return nil
}

func (h *Handler) Pause(guildId uint64) error {
	p, ok := h.m.Get(guildId)
	if !ok {
		return ErrNoActivePlayer
	}

	p.Pause()
	return nil
}

func (h *Handler) Unpause(guildId uint64) error {
	p, ok := h.m.Get(guildId)
	if !ok {
		return ErrNoActivePlayer
	}

	p.Unpause()
	return nil
}

func (h *Handler) EnableLoop(guildId uint64, loop bool) (bool, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return false, ErrNoActivePlayer
	}

	is := p.IsLooping()
	if is == loop {
		return false, nil
	}
	p.SetLoop(loop)

	return true, nil
}

func (h *Handler) SetVolume(guildId uint64, volume uint8) (uint8, error) {
	panic("unimplemented")
}

func (h *Handler) GetTrackById(guildId uint64, id uuid.UUID) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, ErrNoActivePlayer
	}

	t, ok := p.GetById(id)
	if !ok {
		return nil, ErrTrackNotFoundInQueue
	}

	return t, nil
}

func (h *Handler) GetTracks(guildId uint64) ([]player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, ErrNoActivePlayer
	}

	return p.GetQueue(), nil
}

func (h *Handler) RemoveTrack(guildId uint64, id uuid.UUID) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, ErrNoActivePlayer
	}

	t, ok := p.RemoveTrack(id)
	if !ok {
		return nil, ErrTrackNotFoundInQueue
	}

	return t, nil
}
