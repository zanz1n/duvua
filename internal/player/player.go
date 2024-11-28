package player

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	queue   []*player.Track
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
		queue:       []*player.Track{},
		current:     nil,
		paused:      atomic.Bool{},
		mu:          sync.Mutex{},
		Interrupt:   make(chan InterruptType),
	}
}

func (p *GuildPlayer) GetById(uid uuid.UUID) (*player.Track, bool) {
	id := uid.String()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil {
		if p.current.Id == id {
			c := proto.Clone(p.current).(*player.Track)
			return c, true
		}
	}

	for _, t := range p.queue {
		if t.Id == id {
			v := proto.Clone(t).(*player.Track)
			return v, true
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
	return p.Size() == 0
}

func (p *GuildPlayer) AddTrack(track *player.Track) {
	p.mu.Lock()
	p.queue = append(p.queue, track)
	p.mu.Unlock()
}

func (p *GuildPlayer) Pool() *player.Track {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.queue) == 0 {
		p.current = nil
	} else {
		track := p.queue[0]
		p.queue = p.queue[1:]
		p.current = track

		p.current.State = &player.TrackState{
			Progress:     durationpb.New(0),
			PlayingStart: timestamppb.Now(),
		}
	}

	return p.current
}

func (p *GuildPlayer) RemoveById(uid uuid.UUID) (*player.Track, bool) {
	id := uid.String()

	p.mu.Lock()

	if p.current != nil {
		if p.current.Id == id {
			c := proto.Clone(p.current).(*player.Track)
			p.mu.Unlock()

			p.Skip()
			return c, true
		}
	}
	defer p.mu.Unlock()

	index := -1
	track := (*player.Track)(nil)
	for i, t := range p.queue {
		if t.Id == id {
			track, index = t, i
		}
	}

	if 0 > index {
		return nil, false
	}

	p.queue = append(p.queue[:index], p.queue[index+1:]...)
	return track, true
}

func (p *GuildPlayer) RemoveByPosition(pos int) (*player.Track, bool) {
	if pos == 0 {
		v := p.Skip()
		return v, v != nil
	}

	pos--
	p.mu.Lock()
	defer p.mu.Unlock()

	if pos >= len(p.queue) || 0 > pos {
		return nil, false
	}

	track := p.queue[pos]
	p.queue = append(p.queue[:pos], p.queue[pos+1:]...)

	return track, true
}

func (p *GuildPlayer) QueueDuration() time.Duration {
	d := time.Duration(0)

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, track := range p.queue {
		d += track.Data.Duration.AsDuration()
	}

	if p.current != nil {
		d += p.current.Data.Duration.AsDuration()
		if p.current.State != nil {
			d -= atomicLoadDuration(p.current.State)
		}
	}

	return d
}

func (p *GuildPlayer) GetQueue(
	offset, limit int,
) (current *player.Track, tracks []*player.Track, size int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.current != nil {
		current = proto.Clone(p.current).(*player.Track)
	}

	size = len(p.queue)
	if offset >= size || 0 > offset || 0 > limit {
		return
	}

	finish := len(p.queue)
	if size-offset > limit {
		finish = offset + limit
	}

	tracks = make([]*player.Track, finish-offset)
	for i := range finish - offset {
		tracks[i] = p.queue[offset+i]
	}

	return
}

func (p *GuildPlayer) GetMessageChannel() uint64 {
	return p.textChannel.Load()
}

func (p *GuildPlayer) SetMessageChannel(id uint64) {
	p.textChannel.Store(id)
}

func (p *GuildPlayer) IsLooping() bool {
	return p.loop.Load()
}

func (p *GuildPlayer) SetLoop(v bool) {
	p.loop.Store(v)

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current != nil && p.current.State != nil {
		p.current.State.Looping = v
	}
}

func (p *GuildPlayer) Skip() *player.Track {
	p.mu.Lock()

	if p.current == nil {
		p.mu.Unlock()
		return nil
	}
	c := proto.Clone(p.current).(*player.Track)
	p.mu.Unlock()

	p.Interrupt <- InterruptSkip

	return c
}

func (p *GuildPlayer) Stop() {
	p.Interrupt <- InterruptStop
}

func (p *GuildPlayer) Paused() bool {
	return p.paused.Load()
}

func (p *GuildPlayer) Pause() bool {
	if !p.paused.Load() {
		p.paused.Store(true)
		p.Interrupt <- InterruptPause
		return true
	}
	return false
}

func (p *GuildPlayer) Unpause() bool {
	if p.paused.Load() {
		p.paused.Store(false)
		p.Interrupt <- InterruptUnpause
		return true
	}
	return false
}
