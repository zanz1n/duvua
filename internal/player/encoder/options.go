package encoder

import (
	"fmt"

	"github.com/zanz1n/duvua-bot/config"
)

var DefaultEncodeOptions = &EncodeOptions{
	Volume:           256,
	Channels:         2,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          64,
	Mode:             OpusEncodeModeAudio,
	CompressionLevel: 10,
	PacketLoss:       1,
	BufferedFrames:   100,
	StartTime:        0,
	FFmpegPath:       "ffmpeg",
}

func init() {
	cfg := config.GetConfig()
	if cfg.Player.FFmpegExec != "" {
		DefaultEncodeOptions.FFmpegPath = cfg.Player.FFmpegExec
	}
}

type EncodeOptions struct {
	FrameRate        uint32
	Volume           uint16
	Bitrate          uint8
	CompressionLevel uint8
	Channels         uint8
	Mode             EncodeMode
	FrameDuration    FrameDuration
	PacketLoss       uint8
	BufferedFrames   int
	StartTime        int
	FFmpegPath       string
}

type FrameDuration uint8

var _ fmt.Stringer = FrameDuration(0)

const (
	OpusFrameDurationDefault FrameDuration = iota
	OpusFrameDuration20MS
	OpusFrameDuration40MS
	OpusFrameDuration60MS
)

// String implements fmt.Stringer.
func (d FrameDuration) String() string {
	switch d {
	case OpusFrameDuration20MS:
		return "20"
	case OpusFrameDuration40MS:
		return "40"
	case OpusFrameDuration60MS:
		return "60"
	default:
		return "20"
	}
}

type EncodeMode uint8

var _ fmt.Stringer = EncodeMode(0)

const (
	OpusEncodeModeDefault EncodeMode = iota
	OpusEncodeModeVoip
	OpusEncodeModeAudio
	OpusEncodeModeLowDelay
)

// String implements fmt.Stringer.
func (m EncodeMode) String() string {
	switch m {
	case OpusEncodeModeVoip:
		return "voip"
	case OpusEncodeModeAudio:
		return "audio"
	case OpusEncodeModeLowDelay:
		return "lowdelay"
	default:
		return "audio"
	}
}
