package encoder

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonas747/ogg"
	"github.com/zanz1n/duvua/internal/errors"
)

type Session struct {
	opts *EncodeOptions
	r    io.ReadCloser

	ch   chan []byte
	proc *os.Process

	running    atomic.Bool
	frameCount atomic.Uint32

	sync.Mutex
}

var _ io.Closer = &Session{}

func NewSession(r io.ReadCloser, opts *EncodeOptions) *Session {
	if opts == nil {
		opts = DefaultEncodeOptions
	}

	s := &Session{
		opts: opts,
		r:    r,
		ch:   make(chan []byte, opts.BufferedFrames),
	}

	go func() {
		start := time.Now()

		err := s.start()
		took := time.Since(start).Round(time.Millisecond)
		frameCount := s.frameCount.Load()
		if err != nil {
			slog.Warn(
				"Error caught in encoding session",
				"frame_count", frameCount,
				"took", took,
				"error", err,
			)
		} else {
			slog.Debug(
				"Finished encoding session",
				"frame_count", frameCount,
				"took", took,
			)
		}
	}()

	return s
}

func (s *Session) start() error {
	if s.running.Load() {
		return errors.Unexpected("already running")
	}

	defer s.running.Store(false)
	s.running.Store(true)

	args := []string{
		"-stats",
		"-i", "pipe:0",
		"-reconnect", "1",
		"-reconnect_at_eof", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "2",
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
		"-vbr", "on",
		"-compression_level", strconv.Itoa(int(s.opts.CompressionLevel)),
		"-ar", strconv.Itoa(int(s.opts.FrameRate)),
		"-ac", strconv.Itoa(int(s.opts.Channels)),
		"-b:a", strconv.Itoa(int(s.opts.Bitrate) * 1000),
		"-application", s.opts.Mode.String(),
		"-frame_duration", s.opts.FrameDuration.String(),
		"-packet_loss", strconv.Itoa(int(s.opts.PacketLoss)),
		"-threads", "1",
		"-ss", strconv.Itoa(s.opts.StartTime),
		"-filter:a", fmt.Sprintf(
			"volume=%.2f",
			float64(s.opts.Volume)/256.0,
		),
		"pipe:1",
	}

	ffmpeg := exec.Command(s.opts.FFmpegPath, args...)

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return errors.Unexpected("stdout pipe: " + err.Error())
	}
	defer stdout.Close()

	ffmpeg.Stdin = s.r

	// TODO: Handle stderr

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		return errors.Unexpected("stderr pipe: " + err.Error())
	}
	defer stderr.Close()

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			slog.Debug("FFMPEG: " + s.Text())
		}
	}()

	s.Lock()

	if err = ffmpeg.Start(); err != nil {
		s.Unlock()
		return errors.Unexpected("spawn ffmpeg: " + err.Error())
	}
	s.proc = ffmpeg.Process

	s.Unlock()

	slog.Debug(
		"FFmpeg process started",
		"pid", s.proc.Pid,
	)

	defer close(s.ch)
	err = s.readStdout(stdout)
	if err != nil {
		return err
	}

	if err = ffmpeg.Wait(); err != nil {
		if err.Error() == "signal: killed" {
			return nil
		}
		return errors.Unexpected("ffmpeg process: " + err.Error())
	}
	state := ffmpeg.ProcessState

	slog.Debug(
		"FFmpeg process stopped",
		"pid", state.Pid(),
		"exit_code", state.ExitCode(),
		"took", state.UserTime().Round(time.Millisecond),
	)

	return nil
}

func (s *Session) readStdout(r io.Reader) error {
	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(r))

	skipPackets := 2
	for {
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				return errors.Unexpected("ogg decoder: " + err.Error())
			}
			break
		}

		s.ch <- packet
		s.frameCount.Add(1)
	}

	return nil
}

func (s *Session) ReadOpus() ([]byte, error) {
	buf, ok := <-s.ch
	if !ok {
		return nil, io.EOF
	}
	return buf, nil
}

// Close implements io.Closer.
func (s *Session) Close() (err error) {
	if !s.running.Load() {
		return errors.Unexpected("not running")
	}

	s.running.Store(false)

	s.Lock()
	defer s.Unlock()
	if s.proc != nil {
		err = s.proc.Kill()
		s.proc = nil
	} else {
		err = errors.Unexpected("not running")
	}

	s.r.Close()
	s.r = nil

	for range s.ch {
		// cleans the buffered frames
	}
	return
}
