package grpcpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
)

var _ grpc.ClientConnInterface = &Pool{}

type Pool struct {
	connections []*grpc.ClientConn
	mu          sync.Mutex
	next        int

	target string
	opts   []grpc.DialOption
}

func New(size int, target string, opts ...grpc.DialOption) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be greater than zero")
	}

	pool := &Pool{
		connections: make([]*grpc.ClientConn, 0, size),
		target:      target,
		opts:        opts,
	}

	wg := sync.WaitGroup{}
	var someErr atomic.Value

	for range size {
		wg.Add(1)
		go func() {
			if err := pool.AddConnection(); err != nil {
				someErr.Store(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()

	err := someErr.Load()
	if err != nil {
		return nil, err.(error)
	}

	return pool, nil
}

func (p *Pool) AddConnection() error {
	conn, err := grpc.NewClient(p.target, p.opts...)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections = append(p.connections, conn)
	return nil
}

// GetConnection returns the next available connection (round-robin)
func (p *Pool) GetConnection() *grpc.ClientConn {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn := p.connections[p.next]
	p.next = (p.next + 1) % len(p.connections)
	return conn
}

// Close closes all connections in the pool
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.connections {
		conn.Close()
	}
}

// Invoke implements grpc.ClientConnInterface.
func (p *Pool) Invoke(
	ctx context.Context,
	method string,
	args any,
	reply any,
	opts ...grpc.CallOption,
) error {
	return p.GetConnection().Invoke(ctx, method, args, reply, opts...)
}

// NewStream implements grpc.ClientConnInterface.
func (p *Pool) NewStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return p.GetConnection().NewStream(ctx, desc, method, opts...)
}
