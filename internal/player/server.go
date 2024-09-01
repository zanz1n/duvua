package player

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zanz1n/duvua-bot/internal/errors"
)

type HttpServer struct {
	h        *Handler
	authKey  string
	validate *validator.Validate
}

func NewHttpServer(h *Handler, authKey string) *HttpServer {
	return &HttpServer{
		h:        h,
		authKey:  authKey,
		validate: validator.New(),
	}
}

func (s *HttpServer) Route(m *http.ServeMux) {
	// Search for track
	m.HandleFunc("GET /track/search", s.m(s.getTrack))

	// Get current playing track
	m.HandleFunc("GET /guild/{guild_id}/track", s.m(s.getCurrentTrack))
	// Get specific track of the queue
	m.HandleFunc("GET /guild/{guild_id}/track/{id}", s.m(s.getTrackById))
	// Get the entire queue
	m.HandleFunc("GET /guild/{guild_id}/tracks", s.m(s.getAllTracks))

	// Add a track to the queue
	m.HandleFunc("POST /guild/{guild_id}/track", s.m(s.postTrack))

	// Skip track
	m.HandleFunc("PUT /guild/{guild_id}/skip", s.m(s.putSkipTrack))
	// Pause queue
	m.HandleFunc("PUT /guild/{guild_id}/pause", s.m(s.putPauseQueue))
	// Unpause queue
	m.HandleFunc("PUT /guild/{guild_id}/unpause", s.m(s.putUnpauseQueue))
	// Enable/disable track loop
	m.HandleFunc("PUT /guild/{guild_id}/loop", s.m(s.putSetQueueLoop))
	// Set queue volume
	m.HandleFunc("PUT /guild/{guild_id}/volume", s.m(s.putSetQueueVolume))

	// Stop
	m.HandleFunc("DELETE /guild/{guild_id}", s.m(s.deleteQueue))
	// Remove track
	m.HandleFunc("DELETE /guild/{guild_id}/track/{id}", s.m(s.deleteTrackById))
}

type handlerFunc = func(w http.ResponseWriter, r *http.Request) error

func (s *HttpServer) m(h handlerFunc) http.HandlerFunc {
	return s.loggerMiddleware(
		s.errorMiddleware(
			s.catchPanicMiddleware(
				s.authMiddleware(h),
			),
		),
	)
}

func (s *HttpServer) authMiddleware(h handlerFunc) handlerFunc {
	if s.authKey != "" {
		ak := "passwd:" + s.authKey

		return func(w http.ResponseWriter, r *http.Request) error {
			authH := r.Header.Get("Authorization")
			if authH != ak {
				w.WriteHeader(http.StatusUnauthorized)
				return errors.New("auth required with `Authorization` header")
			}

			return h(w, r)
		}
	}
	return h
}

func (s *HttpServer) catchPanicMiddleware(h handlerFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		defer func() {
			if err := recover(); err != nil {
				handleErrRes(w, err)
			}
		}()

		return h(w, r)
	}
}

func (s *HttpServer) errorMiddleware(h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			handleErrRes(w, err)
		}
	}
}

func (s *HttpServer) loggerMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w2 := wrapResponseWriter(w)
		h(w2, r)
		slog.Info(
			"HTTP: Incomming request",
			"path", r.URL.Path,
			"status_code", w2.Status(),
			"took", time.Since(start).Round(10*time.Microsecond),
		)
	}
}

func (s *HttpServer) parseReqBody(body io.Reader, v any) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return errors.New("failed to read request body")
	}

	if err = json.Unmarshal(buf, v); err != nil {
		return errors.New("failed to unmarshal request body: " + err.Error())
	}
	if err = s.validate.Struct(v); err != nil {
		return errors.New("invalid request body: " + err.Error())
	}

	return nil
}

func getUuidPathParam(r *http.Request, name string) (uuid.UUID, error) {
	pv := r.PathValue(name)
	if pv != "" {
		return uuid.Nil, errors.Newf("path parameter `%s` is required", name)
	}

	v, err := uuid.Parse(pv)
	if err != nil {
		return uuid.Nil, errors.Newf("invalid path parameter `%s`", name)
	}

	return v, nil
}

func getUintPathParam(r *http.Request, name string) (uint64, error) {
	pv := r.PathValue(name)
	if pv != "" {
		return 0, errors.Newf("path parameter `%s` is required", name)
	}

	v, err := strconv.ParseUint(pv, 10, 0)
	if err != nil {
		return 0, errors.Newf("invalid path parameter `%s`", name)
	}

	return v, nil
}

func handleErrRes(w http.ResponseWriter, err any) {
	const DefaultErr = `{"error":"something went wrong","error_code":0}`

	type errStruct struct {
		Error     string `json:"error"`
		ErrorCode int    `json:"error_code"`
	}

	es := errStruct{Error: fmt.Sprint(err)}
	if e, ok := err.(error); ok {
		es.ErrorCode = int(errToErrCode(e))
	}

	w.Header().Add("Content-Type", "application/json")

	buf, err := json.Marshal(es)
	if err != nil {
		w.Write([]byte(DefaultErr))
	} else {
		w.Write(buf)
	}
}

func resJson[T any](w http.ResponseWriter, data T, message string) error {
	type dataBody[T any] struct {
		Message string `json:"message"`
		Data    T      `json:"data"`
	}

	b, err := json.Marshal(dataBody[T]{
		Message: message,
		Data:    data,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return errors.Unexpected("failed to marshal response: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)

	return nil
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}
