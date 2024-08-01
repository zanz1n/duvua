package kv

import (
	enc "encoding/gob"
	"io"
	"log/slog"
	"os"
	"time"
)

func (m *MemoryKeyValueRepository) Init() {
	go m.BackgroundJob()
}

func (m *MemoryKeyValueRepository) BackgroundJob() {
	ticker := time.NewTicker(m.saveTimeout)

	for {
		<-ticker.C

		m.ClearDangling()
		m.Save()
	}
}

func (m *MemoryKeyValueRepository) ClearDangling() {
	start := time.Now()
	toRemove := []string{}

	m.mpMu.RLock()
	for k, v := range m.mp {
		if v.IsExpired() {
			toRemove = append(toRemove, k)
		}
	}
	m.mpMu.RUnlock()

	if len(toRemove) > 0 {
		m.mpMu.Lock()
		for _, k := range toRemove {
			delete(m.mp, k)
		}
		m.mpMu.Unlock()

		slog.Info(
			"Deleted dangling key-value registries",
			"amount", len(toRemove),
			"took", time.Since(start),
		)
	} else {
		slog.Info(
			"No dangling key-value registries found",
			"took", time.Since(start),
		)
	}
}

func (m *MemoryKeyValueRepository) ReadContent(r io.Reader) error {
	dec := enc.NewDecoder(r)

	tempMp := make(map[string]inMemoryValue)
	if err := dec.Decode(&tempMp); err != nil {
		return err
	}

	if len(tempMp) > 0 {
		m.mpMu.Lock()
		for k, v := range tempMp {
			m.mp[k] = v
		}
		m.mpMu.Unlock()
	}

	return nil
}

func (m *MemoryKeyValueRepository) WriteContent(w io.Writer) error {
	enc := enc.NewEncoder(w)

	m.mpMu.Lock()
	err := enc.Encode(m.mp)
	m.mpMu.Unlock()

	return err
}

func (m *MemoryKeyValueRepository) Load() error {
	if m.filePath == nil {
		return ErrFilePathNotProvided
	}

	file, err := os.OpenFile(*m.filePath, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = m.ReadContent(file); err == io.EOF {
		return nil
	}
	return err
}

func (m *MemoryKeyValueRepository) save() error {
	if m.filePath == nil {
		return ErrFilePathNotProvided
	}

	file, err := os.OpenFile(*m.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	err = m.WriteContent(file)
	return err
}

func (m *MemoryKeyValueRepository) Save() {
	start := time.Now()

	if err := m.save(); err != nil {
		if err != ErrFilePathNotProvided {
			slog.Error(
				"Failed to save key-value data to file",
				"took", time.Since(start),
				"error", err,
			)
		} else {
			slog.Info("Skipped key-value save: file path not provided")
		}

	} else {
		slog.Info(
			"Saved key-value data",
			"file", *m.filePath,
			"took", time.Since(start),
		)
	}
}
