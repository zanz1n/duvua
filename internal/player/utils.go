package player

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/protobuf/types/known/durationpb"
)

func atomicAddDuration(state *player.TrackState, inc time.Duration) {
	ptr := unsafe.Pointer(&state.Progress)
	old := (*durationpb.Duration)(atomic.LoadPointer(&ptr)).AsDuration()
	new := durationpb.New(old + inc)
	atomic.StorePointer(&ptr, unsafe.Pointer(&new))
}

func atomicLoadDuration(state *player.TrackState) time.Duration {
	ptr := unsafe.Pointer(&state.Progress)
	duration := (*durationpb.Duration)(atomic.LoadPointer(&ptr)).AsDuration()
	return duration
}
