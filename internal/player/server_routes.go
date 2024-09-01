package player

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/pkg/player"
)

func (s *HttpServer) getTrack(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	q := r.URL.Query()
	query := q.Get("query")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("query parameter `query` is necessary")
	}

	data, err := s.h.FetchTrack(query)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, data, "success")
}

func (s *HttpServer) getCurrentTrack(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	track, err := s.h.GetPlayingTrack(guildId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, track, "track found")
}

func (s *HttpServer) getTrackById(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	trackId, err := getUuidPathParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	track, err := s.h.GetTrackById(guildId, trackId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, track, "found track")
}

func (s *HttpServer) getAllTracks(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	tracks, err := s.h.GetTracks(guildId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	msg := fmt.Sprintf("%d tracks in queue", len(tracks))
	return resJson(w, tracks, msg)
}

func (s *HttpServer) postTrack(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	var data player.AddTrackData
	if err = s.parseReqBody(r.Body, &data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	track := s.h.AddTrack(guildId, data)
	return resJson(w, track, "added track to queue")
}

func (s *HttpServer) putSkipTrack(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	track, err := s.h.Skip(guildId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}
	return resJson(w, track, "skipped track")
}

func (s *HttpServer) putPauseQueue(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if err = s.h.Pause(guildId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, (*struct{})(nil), "paused queue")
}

func (s *HttpServer) putUnpauseQueue(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if err = s.h.Unpause(guildId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, (*struct{})(nil), "unpaused queue")
}

func (s *HttpServer) putSetQueueLoop(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	q := r.URL.Query()
	enable, err := strconv.ParseBool(q.Get("enable"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New(
			"query parameter `enable` must be a valid bool",
		)
	}

	beforeEnable, err := s.h.EnableLoop(guildId, enable)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	msg := "loop "
	if beforeEnable == enable {
		msg += "already "
	}
	if enable {
		msg += "enabled"
	} else {
		msg += "disabled"
	}

	return resJson(w, (*struct{})(nil), msg)
}

func (s *HttpServer) putSetQueueVolume(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	q := r.URL.Query()
	volume, err := strconv.Atoi(q.Get("volume"))
	if err != nil || 0 > volume || volume > 255 {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New(
			"query parameter `volume` must be a valid uint8",
		)
	}

	beforeVolume, err := s.h.SetVolume(guildId, uint8(volume))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	msg := fmt.Sprintf("volume set from %d to %d", beforeVolume, volume)
	return resJson(w, (*struct{})(nil), msg)
}

func (s *HttpServer) deleteQueue(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if err = s.h.Stop(guildId); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, (*struct{})(nil), "queue stopped")
}

func (s *HttpServer) deleteTrackById(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	trackId, err := getUuidPathParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	track, err := s.h.RemoveTrack(guildId, trackId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, track, "track removed")
}
