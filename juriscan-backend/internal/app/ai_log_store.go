package app

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
)

type AILogStore struct {
	mu    sync.RWMutex
	items []AILogRecord
}

func NewAILogStore() *AILogStore {
	return &AILogStore{
		items: make([]AILogRecord, 0, 128),
	}
}

func (s *AILogStore) Add(record AILogRecord) AILogRecord {
	now := time.Now().UTC()
	item := AILogRecord{
		ID:             newAILogID(),
		Feature:        strings.TrimSpace(record.Feature),
		ResourceType:   strings.TrimSpace(record.ResourceType),
		ResourceID:     strings.TrimSpace(record.ResourceID),
		Prompt:         strings.TrimSpace(record.Prompt),
		Response:       strings.TrimSpace(record.Response),
		Model:          strings.TrimSpace(record.Model),
		Confidence:     record.Confidence,
		CreatedAt:      now,
		RetentionUntil: record.RetentionUntil.UTC(),
	}
	if item.RetentionUntil.IsZero() {
		item.RetentionUntil = now.AddDate(0, 0, 90)
	}

	s.mu.Lock()
	s.items = append(s.items, item)
	s.mu.Unlock()

	return item
}

func (s *AILogStore) List(feature string) []AILogRecord {
	feature = strings.TrimSpace(strings.ToLower(feature))
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]AILogRecord, 0, len(s.items))
	for _, item := range s.items {
		if feature != "" && strings.ToLower(item.Feature) != feature {
			continue
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *AILogStore) PurgeExpired(now time.Time) int {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	keep := s.items[:0]
	removed := 0
	for _, item := range s.items {
		if item.RetentionUntil.Before(now) {
			removed++
			continue
		}
		keep = append(keep, item)
	}
	s.items = keep
	return removed
}

func newAILogID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "ailog-fallback"
	}
	return "ailog-" + hex.EncodeToString(buf)
}
