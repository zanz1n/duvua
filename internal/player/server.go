package player

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zanz1n/duvua/internal/player/errcodes"
	"github.com/zanz1n/duvua/internal/player/platform"
	"github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
	m *PlayerManager
	f *platform.Fetcher
	player.UnimplementedPlayerServer
}

func NewGrpcServer(manager *PlayerManager, f *platform.Fetcher) *GrpcServer {
	return &GrpcServer{m: manager, f: f}
}

// Add implements player.PlayerServer.
func (s *GrpcServer) Add(
	ctx context.Context,
	req *player.AddRequest,
) (*player.AddResponse, error) {
	p := s.m.GetOrCreate(req.GuildId, req.ChannelId)

	tracks := make([]*player.Track, len(req.Data))
	for i, track := range req.Data {
		track := &player.Track{
			Id:        uuid.NewString(),
			CreatedAt: timestamppb.Now(),
			UserId:    req.UserId,
			ChannelId: req.ChannelId,
			State:     nil,
			Data:      track,
		}
		p.AddTrack(track)

		tracks[i] = track

		slog.Info(
			"Added track to queue",
			"guild_id", req.GuildId,
			"user_id", req.UserId,
			"url", track.Data.PlayQuery,
			"duration", track.Data.Duration.AsDuration().Round(time.Millisecond),
		)
	}

	p.SetMessageChannel(req.TextChannelId)

	return &player.AddResponse{Tracks: tracks}, nil
}

// EnableLoop implements player.PlayerServer.
func (s *GrpcServer) EnableLoop(
	ctx context.Context,
	req *player.EnableLoopRequest,
) (*player.ChangedResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	is := p.IsLooping()
	if is == req.Enable {
		return &player.ChangedResponse{Changed: false}, nil
	}

	p.SetLoop(req.Enable)

	return &player.ChangedResponse{Changed: false}, nil
}

// Fetch implements player.PlayerServer.
func (s *GrpcServer) Fetch(
	ctx context.Context,
	req *player.FetchRequest,
) (*player.FetchResponse, error) {
	data, err := s.f.Search(req.Query)
	if err != nil {
		return nil, err
	}

	return &player.FetchResponse{Data: data}, nil
}

// GetAll implements player.PlayerServer.
func (s *GrpcServer) GetAll(
	ctx context.Context,
	req *player.GetAllRequest,
) (*player.GetAllResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	d := p.QueueDuration()
	playing, tracks, totalsize := p.GetQueue(int(req.Offset), int(req.Limit))

	return &player.GetAllResponse{
		TotalSize:     int32(totalsize),
		TotalDuration: durationpb.New(d),
		Playing:       playing,
		Tracks:        tracks,
	}, nil
}

// GetById implements player.PlayerServer.
func (s *GrpcServer) GetById(
	ctx context.Context,
	req *player.TrackIdRequest,
) (*player.TrackResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	id, _ := uuid.Parse(req.Id)
	track, ok := p.GetById(id)
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return &player.TrackResponse{Track: track}, nil
}

// GetCurrent implements player.PlayerServer.
func (s *GrpcServer) GetCurrent(
	ctx context.Context,
	req *player.GuildIdRequest,
) (*player.TrackResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	track, ok := p.GetCurrent()
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return &player.TrackResponse{Track: track}, nil
}

// Pause implements player.PlayerServer.
func (s *GrpcServer) Pause(
	ctx context.Context,
	req *player.GuildIdRequest,
) (*player.ChangedResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	changed := p.Pause()

	return &player.ChangedResponse{Changed: changed}, nil
}

// Remove implements player.PlayerServer.
func (s *GrpcServer) Remove(
	ctx context.Context,
	req *player.TrackIdRequest,
) (*player.TrackResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	id, _ := uuid.Parse(req.Id)
	track, ok := p.RemoveById(id)
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return &player.TrackResponse{Track: track}, nil
}

// RemoveByPosition implements player.PlayerServer.
func (s *GrpcServer) RemoveByPosition(
	ctx context.Context,
	req *player.RemoveByPositionRequest,
) (*player.TrackResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	track, ok := p.RemoveByPosition(int(req.Position))
	if !ok {
		return nil, errcodes.ErrTrackNotFoundInQueue
	}

	return &player.TrackResponse{Track: track}, nil
}

// SetVolume implements player.PlayerServer.
func (s *GrpcServer) SetVolume(
	ctx context.Context,
	req *player.SetVolumeRequest,
) (*player.ChangedResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

// Skip implements player.PlayerServer.
func (s *GrpcServer) Skip(
	ctx context.Context,
	req *player.GuildIdRequest,
) (*player.TrackResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	track := p.Skip()
	if track == nil {
		return nil, errcodes.ErrNoActivePlayer
	}

	return &player.TrackResponse{Track: track}, nil
}

// Stop implements player.PlayerServer.
func (s *GrpcServer) Stop(
	ctx context.Context,
	req *player.GuildIdRequest,
) (*emptypb.Empty, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	p.Stop()
	return &emptypb.Empty{}, nil
}

// Unpause implements player.PlayerServer.
func (s *GrpcServer) Unpause(
	ctx context.Context,
	req *player.GuildIdRequest,
) (*player.ChangedResponse, error) {
	p, ok := s.m.Get(req.GuildId)
	if !ok {
		return nil, errcodes.ErrNoActivePlayer
	}

	changed := p.Unpause()

	return &player.ChangedResponse{Changed: changed}, nil
}
