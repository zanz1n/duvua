package kv_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zanz1n/duvua-bot/internal/kv"
)

func TestSetValue(t *testing.T) {
	r := kv.NewMemoryKeyValue(60 * time.Second)

	key := uuid.NewString()
	value := uuid.NewString()

	assert.Nil(t, r.Set(key, value), "Failed to Set on kv")

	storedV, err := r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, value, storedV, "Stored value mismatches the setted one")
}

func TestDeleteValue(t *testing.T) {
	r := kv.NewMemoryKeyValue(60 * time.Second)

	key := uuid.NewString()
	value := uuid.NewString()

	assert.Nil(t, r.Set(key, value), "Failed to Set on kv")

	assert.Nil(t, r.Delete(key), "Failed to delete kv")

	storedV, err := r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, "", storedV, "Retrieved value successfully after its exclusion")
}

func TestGetDeleteValue(t *testing.T) {
	r := kv.NewMemoryKeyValue(60 * time.Second)

	key := uuid.NewString()
	value := uuid.NewString()

	assert.Nil(t, r.Set(key, value), "Failed to Set on kv")

	storedV, err := r.GetDelete(key)
	assert.Nil(t, err, "Failed to delete kv")
	assert.Equal(t, value, storedV, "Stored value mismatches the setted one")

	storedV, err = r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, "", storedV, "Retrieved value successfully after its exclusion")
}

func TestSetTtlValue(t *testing.T) {
	r := kv.NewMemoryKeyValue(60 * time.Second)

	key := uuid.NewString()
	value := uuid.NewString()
	ttl := 100 * time.Millisecond

	assert.Nil(t, r.SetTtl(key, value, ttl), "Failed to SetTtl on kv")

	storedV, err := r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, value, storedV, "Stored value mismatches the setted one")

	time.Sleep(2 * ttl)

	storedV, err = r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, "", storedV, "Value was not expired properly after the end of TTL")
}

func TestGetTtlValue(t *testing.T) {
	r := kv.NewMemoryKeyValue(60 * time.Second)

	key := uuid.NewString()
	value := uuid.NewString()
	ttl := 100 * time.Millisecond

	assert.Nil(t, r.Set(key, value), "Failed to Set on kv")

	storedV, err := r.GetTtl(key, ttl)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, value, storedV, "Stored value mismatches the setted one")

	time.Sleep(2 * ttl)

	storedV, err = r.Get(key)
	assert.Nil(t, err, "Failed to retrieve stored value")
	assert.Equal(t, "", storedV, "Value was not expired properly after the end of TTL")
}
