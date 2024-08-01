package kv

import (
	"sync"
	"time"
)

type inMemoryValue struct {
	Expiry *time.Time
	Value  any
}

func (v *inMemoryValue) IsExpired() bool {
	if v.Expiry == nil {
		return false
	}
	return time.Now().UnixMilli() > v.Expiry.UnixMilli()
}

func NewMemoryKeyValue(saveTimeout time.Duration, filePath ...string) *MemoryKeyValueRepository {
	var fpath *string = nil
	if len(filePath) > 0 {
		fpath = &filePath[0]
	}

	return &MemoryKeyValueRepository{
		mp:          make(map[string]inMemoryValue),
		mpMu:        sync.RWMutex{},
		saveTimeout: saveTimeout,
		filePath:    fpath,
	}
}

type MemoryKeyValueRepository struct {
	mp          map[string]inMemoryValue
	mpMu        sync.RWMutex
	saveTimeout time.Duration
	filePath    *string
}

func (m *MemoryKeyValueRepository) delete(key string) any {
	m.mpMu.RLock()
	if v, ok := m.mp[key]; ok {
		m.mpMu.RUnlock()

		m.mpMu.Lock()
		delete(m.mp, key)
		m.mpMu.Unlock()

		if v.IsExpired() {
			return nil
		}
		return v.Value
	}
	m.mpMu.RUnlock()

	return nil
}

func (m *MemoryKeyValueRepository) get(key string, ttl *time.Duration) any {
	m.mpMu.RLock()

	if v, ok := m.mp[key]; ok {
		if v.IsExpired() {
			m.mpMu.RUnlock()
			return m.delete(key)
		}
		m.mpMu.RUnlock()

		if ttl != nil {
			expiry := time.Now().Add(*ttl)

			m.mpMu.Lock()
			m.mp[key] = inMemoryValue{
				Value:  v.Value,
				Expiry: &expiry,
			}
			m.mpMu.Unlock()
		}

		return v.Value
	}
	m.mpMu.RUnlock()

	return nil
}

func (m *MemoryKeyValueRepository) set(key string, ttl *time.Duration, value any) {
	var expiry *time.Time = nil
	if ttl != nil {
		e := time.Now().Add(*ttl)
		expiry = &e
	}

	m.mpMu.Lock()
	defer m.mpMu.Unlock()

	m.mp[key] = inMemoryValue{
		Value:  value,
		Expiry: expiry,
	}
}

// Get implements BasicKeyValue.
func (m *MemoryKeyValueRepository) Get(key string) (string, error) {
	value := m.get(key, nil)
	if value == nil {
		return "", nil
	}

	if s, ok := value.(string); ok {
		return s, nil
	}

	return "", ErrNotStringValue
}

// GetTtl implements BasicKeyValue.
func (m *MemoryKeyValueRepository) GetTtl(key string, ttl time.Duration) (string, error) {
	value := m.get(key, &ttl)
	if value == nil {
		return "", nil
	}

	if s, ok := value.(string); ok {
		return s, nil
	}
	return "", ErrNotStringValue
}

// Set implements BasicKeyValue.
func (m *MemoryKeyValueRepository) Set(key string, value string) error {
	m.set(key, nil, value)
	return nil
}

// SetTtl implements BasicKeyValue.
func (m *MemoryKeyValueRepository) SetTtl(key string, value string, ttl time.Duration) error {
	m.set(key, &ttl, value)
	return nil
}

// Delete implements BasicKeyValue.
func (m *MemoryKeyValueRepository) Delete(key string) error {
	m.delete(key)
	return nil
}

// GetDelete implements BasicKeyValue.
func (m *MemoryKeyValueRepository) GetDelete(key string) (string, error) {
	value := m.delete(key)
	if value == nil {
		return "", nil
	}

	if s, ok := value.(string); ok {
		return s, nil
	}
	return "", ErrNotStringValue
}

// GetUnmarshal implements SerializableKeyValue.
func (m *MemoryKeyValueRepository) GetUnmarshal(key string, v any) (bool, error) {
	panic("unimplemented")
}

// GetUnmarshalTtl implements SerializableKeyValue.
func (m *MemoryKeyValueRepository) GetUnmarshalTtl(key string, ttl time.Duration, v any) (bool, error) {
	panic("unimplemented")
}

// SetMarshal implements SerializableKeyValue.
func (m *MemoryKeyValueRepository) SetMarshal(key string, value any) error {
	m.set(key, nil, value)
	return nil
}

// SetMarshalTtl implements SerializableKeyValue.
func (m *MemoryKeyValueRepository) SetMarshalTtl(key string, value any, ttl time.Duration) error {
	m.set(key, &ttl, value)
	return nil
}

// GetDeleteUnmarshal implements SerializableKeyValue.
func (m *MemoryKeyValueRepository) GetDeleteUnmarshal(key string, v any) (bool, error) {
	panic("unimplemented")
}
