package player

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/internal/player/platform"
	"github.com/zanz1n/duvua/pkg/player"
)

type Handler struct {
	m *PlayerManager
	f *platform.Fetcher
}

func NewHandler(manager *PlayerManager, f *platform.Fetcher) *Handler {
	return &Handler{m: manager, f: f}
}

func (h *Handler) FetchTrack(query string) ([]player.TrackData, error) {
	return h.f.Search(query)
}

func (h *Handler) AddTrack(guildId uint64, data player.AddTrackData) ([]player.Track, error) {
	userId, _ := atoi(data.UserId)
	channelId, _ := atoi(data.ChannelId)
	textChannelId, _ := atoi(data.TextChannelId)

	if userId == 0 || channelId == 0 || textChannelId == 0 {
		return nil, errors.New(
			"`user_id` and `channel_id` must be valid uint64",
		)
	}

	p := h.m.GetOrCreate(guildId, channelId)

	tracks := make([]player.Track, len(data.Data))
	for i, track := range data.Data {
		track := player.NewTrack(userId, channelId, &track)
		p.AddTrack(track)

		tracks[i] = track

		slog.Info(
			"Added track to queue",
			"guild_id", guildId,
			"user_id", data.UserId,
			"url", track.Data.PlayQuery,
			"duration", track.Data.Duration.Round(time.Millisecond),
		)
	}

	p.SetMessageChannel(textChannelId)

	return tracks, nil
}

func (h *Handler) GetPlayingTrack(guildId uint64) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	current, ok := p.GetCurrent()
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}
	return current, nil
}

func (h *Handler) Skip(guildId uint64) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	t := p.Skip()
	if t == nil {
		return nil, errcodes.ErrNoActivePlayer
	}

	return t, nil
}

func (h *Handler) Stop(guildId uint64) error {
	p, ok := h.m.Get(guildId)
	if !ok {
		return errcodes.ErrNoActivePlayer
	}

	p.Stop()
	return nil
}

func (h *Handler) Pause(guildId uint64) (bool, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return false, errcodes.ErrNoActivePlayer
	}

	return p.Pause(), nil
}

func (h *Handler) Unpause(guildId uint64) (bool, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return false, errcodes.ErrNoActivePlayer
	}

	return p.Unpause(), nil
}

func (h *Handler) EnableLoop(guildId uint64, loop bool) (bool, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return false, errcodes.ErrNoActivePlayer
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
		return nil, errcodes.ErrNoActivePlayer
	}

	t, ok := p.GetById(id)
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return t, nil
}

func (h *Handler) GetTracks(guildId uint64, offset int, limit int) (*player.GetTracksData, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	d := p.QueueDuration()
	playing, tracks, totalsize := p.GetQueue(offset, limit)

	data := player.GetTracksData{
		TotalSize:     totalsize,
		TotalDuration: d,
		Playing:       playing,
		Tracks:        tracks,
	}

	return &data, nil
}

func (h *Handler) RemoveTrack(guildId uint64, id uuid.UUID) (*player.Track, error) {
	p, ok := h.m.Get(guildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	t, ok := p.RemoveTrack(id)
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return t, nil
}

func atoi(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 0)
}
