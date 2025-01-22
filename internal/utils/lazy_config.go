package utils

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func NewLazyConfig[T any]() LazyConfig[T] {
	return LazyConfig[T]{lazy: NewLazy(func() (*T, error) {
		godotenv.Load()
		var instance T

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := envconfig.Process(ctx, &instance); err != nil {
			return nil, err
		}
		return &instance, nil
	})}
}

type LazyConfig[T any] struct {
	lazy Lazy[T]
}

func (l *LazyConfig[T]) Get() *T {
	v, err := l.lazy.Get()
	if err != nil {
		log.Fatalln("Failed to init configuration:", err)
	}
	return v
}

func (l *LazyConfig[T]) Init() error {
	_, err := l.lazy.Get()
	return err
}
