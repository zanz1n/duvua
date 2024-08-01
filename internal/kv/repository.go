package kv

import (
	"time"
)

type BasicKeyValue interface {
	Get(key string) (string, error)
	GetTtl(key string, ttl time.Duration) (string, error)

	Set(key string, value string) error
	SetTtl(key string, value string, ttl time.Duration) error

	Delete(key string) error
	GetDelete(key string) (string, error)
}

type SerializableKeyValue interface {
	GetUnmarshal(key string, v any) (bool, error)
	GetUnmarshalTtl(key string, ttl time.Duration, v any) (bool, error)

	SetMarshal(key string, value any) error
	SetMarshalTtl(key string, value any, ttl time.Duration) error

	GetDeleteUnmarshal(key string, v any) (bool, error)
}

type KeyValueRepository interface {
	BasicKeyValue
	SerializableKeyValue
}
