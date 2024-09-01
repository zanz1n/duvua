package player

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/zanz1n/duvua-bot/pkg/player"
)

type InterruptType uint8

const (
	InterruptNone InterruptType = iota
	InterruptSkip
	InterruptStop
	InterruptPause
	InterruptUnpause
	// InterruptSetVolume
	// InterruptSetSpeed
)

var _ fmt.Stringer = InterruptNone

// String implements fmt.Stringer.
func (i InterruptType) String() string {
	switch i {
	case InterruptSkip:
		return "skip"
	case InterruptStop:
		return "stop"
	case InterruptPause:
		return "pause"
	case InterruptUnpause:
		return "unpause"
	// case InterruptSetVolume:
	// 	return "set-volume"
	// case InterruptSetSpeed:
	// 	return "set-speed"
	default:
		return "none"
	}
}

type GuildPlayer struct {
	GuildId     uint64
	loop        atomic.Bool
	textChannel atomic.Uint64

	queue   []player.Track
	current *player.Track
	paused  atomic.Bool

	mu sync.Mutex

	Interrupt chan InterruptType
}

func newGuildPlayer(guildId uint64) *GuildPlayer {
	return &GuildPlayer{
		GuildId:     guildId,
		loop:        atomic.Bool{},
		textChannel: atomic.Uint64{},
		queue:       []player.Track{},
		current:     nil,
		paused:      atomic.Bool{},
		mu:          sync.Mutex{},
		Interrupt:   make(chan InterruptType),
	}
}

func (p *GuildPlayer) GetById(id uuid.UUID) (*player.Track, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil {
		if p.current.ID == id {
			c := *p.current
			return &c, true
		}
	}

	for _, t := range p.queue {
		if t.ID == id {
			v := t
			return &v, true
		}
	}

	return nil, false
}

func (p *GuildPlayer) GetCurrent() (*player.Track, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current == nil {
		return nil, false
	}
	return p.current, true
}

func (p *GuildPlayer) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.queue)
}

func (p *GuildPlayer) IsEmpty() bool {
	empty := p.Size() == 0

	return empty
}

func (p *GuildPlayer) AddTrack(track player.Track) {
	p.mu.Lock()
	p.queue = append(p.queue, track)
	p.mu.Unlock()
}

func (p *GuildPlayer) Pool() *player.Track {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.queue) == 0 {
		return nil
	}

	track := p.queue[0]
	p.queue = p.queue[1:]
	p.current = &track

	p.current.State = &player.TrackState{
		Progress:     0,
		PlayingStart: time.Now(),
	}

	return p.current
}

func (p *GuildPlayer) RemoveTrack(id uuid.UUID) (*player.Track, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	index := -1
	track := player.Track{}
	for i, t := range p.queue {
		if t.ID == id {
			track, index = t, i
		}
	}

	if 0 > index {
		return nil, false
	}

	p.queue = append(p.queue[:index], p.queue[index+1:]...)
	return &track, true
}

func (p *GuildPlayer) GetQueue() []player.Track {
	p.mu.Lock()
	defer p.mu.Unlock()

	dst := make([]player.Track, len(p.queue))
	copy(dst, p.queue)

	return dst
}

func (p *GuildPlayer) SetMessageChannel(id uint64) {
	p.textChannel.Store(id)
}

func (p *GuildPlayer) IsLooping() bool {
	return p.loop.Load()
}

func (p *GuildPlayer) SetLoop(v bool) {
	p.loop.Store(v)
}

func (p *GuildPlayer) Skip() *player.Track {
	if p.current == nil {
		return nil
	}
	p.mu.Lock()
	c := *p.current
	p.mu.Unlock()

	p.Interrupt <- InterruptSkip

	return &c
}

func (p *GuildPlayer) Stop() {
	p.Interrupt <- InterruptStop
}

func (p *GuildPlayer) Paused() bool {
	return p.paused.Load()
}

func (p *GuildPlayer) Pause() {
	if !p.paused.Load() {
		p.paused.Store(true)
		p.Interrupt <- InterruptPause
	}
}

func (p *GuildPlayer) Unpause() {
	if p.paused.Load() {
		p.paused.Store(false)
		p.Interrupt <- InterruptUnpause
	}
}
