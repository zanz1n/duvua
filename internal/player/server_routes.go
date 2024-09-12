package player

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/pkg/player"
)

var NILD = (*struct{})(nil)

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

	return resJson(w, data, "success", false)
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

	return resJson(w, track, "track found", false)
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

	return resJson(w, track, "found track", false)
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
	return resJson(w, tracks, msg, false)
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

	track, err := s.h.AddTrack(guildId, data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	return resJson(w, track, "added track to queue", true)
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
	return resJson(w, track, "skipped track", track != nil)
}

func (s *HttpServer) putPauseQueue(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	changed, err := s.h.Pause(guildId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, NILD, "paused queue", changed)
}

func (s *HttpServer) putUnpauseQueue(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	guildId, err := getUintPathParam(r, "guild_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	changed, err := s.h.Unpause(guildId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	return resJson(w, NILD, "unpaused queue", changed)
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

	changed, err := s.h.EnableLoop(guildId, enable)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	msg := "loop "
	if !changed {
		msg += "already "
	}
	if enable {
		msg += "enabled"
	} else {
		msg += "disabled"
	}

	return resJson(w, NILD, msg, changed)
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

	newVolume := uint8(volume)
	beforeVolume, err := s.h.SetVolume(guildId, newVolume)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	msg := fmt.Sprintf("volume set from %d to %d", beforeVolume, volume)
	return resJson(w, NILD, msg, beforeVolume != newVolume)
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

	return resJson(w, NILD, "queue stopped and deleted", true)
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

	return resJson(w, track, "track removed", track != nil)
}
