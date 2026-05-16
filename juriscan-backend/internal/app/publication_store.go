package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrPublicationNotFound = errors.New("publication: not found")
	ErrInvalidPublication  = errors.New("publication: invalid payload")
	ErrPublicationNotReady = errors.New("publication: analysis required")
)

type PublicationStore struct {
	mu    sync.RWMutex
	items []Publication
}

func NewPublicationStore() *PublicationStore {
	return &PublicationStore{
		items: make([]Publication, 0, 64),
	}
}

func (s *PublicationStore) List() []Publication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := append([]Publication(nil), s.items...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *PublicationStore) Create(input Publication) (Publication, error) {
	source := strings.TrimSpace(input.Source)
	inputType := strings.TrimSpace(strings.ToLower(input.InputType))
	rawText := strings.TrimSpace(input.RawText)
	createdBy := strings.TrimSpace(strings.ToLower(input.CreatedBy))
	fileName := strings.TrimSpace(input.FileName)
	if source == "" || rawText == "" || createdBy == "" {
		return Publication{}, ErrInvalidPublication
	}
	if inputType == "" {
		inputType = "texto"
	}
	if inputType != "texto" && inputType != "arquivo" {
		return Publication{}, ErrInvalidPublication
	}

	now := time.Now().UTC()
	item := Publication{
		ID:        newPublicationID(),
		Source:    source,
		InputType: inputType,
		FileName:  fileName,
		RawText:   rawText,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.mu.Lock()
	s.items = append(s.items, item)
	s.mu.Unlock()

	return item, nil
}

func (s *PublicationStore) GetByID(id string) (Publication, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Publication{}, ErrInvalidPublication
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}
	return Publication{}, ErrPublicationNotFound
}

func (s *PublicationStore) SetAnalysis(id string, analysis PublicationAnalysis) (Publication, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Publication{}, ErrInvalidPublication
	}
	analysis.ActType = strings.TrimSpace(strings.ToLower(analysis.ActType))
	analysis.Risk = strings.TrimSpace(strings.ToLower(analysis.Risk))
	analysis.SuggestedOwnerEmail = strings.TrimSpace(strings.ToLower(analysis.SuggestedOwnerEmail))
	analysis.Prompt = strings.TrimSpace(analysis.Prompt)
	analysis.Response = strings.TrimSpace(analysis.Response)
	analysis.Model = strings.TrimSpace(analysis.Model)
	if analysis.ActType == "" || analysis.Risk == "" || analysis.SuggestedDeadlineDays <= 0 || analysis.SuggestedOwnerEmail == "" {
		return Publication{}, ErrInvalidPublication
	}
	if analysis.BaseDate.IsZero() {
		analysis.BaseDate = time.Now().UTC()
	}
	if analysis.SuggestedDeadlineAt.IsZero() {
		analysis.SuggestedDeadlineAt = analysis.BaseDate.AddDate(0, 0, analysis.SuggestedDeadlineDays)
	}
	if analysis.GeneratedAt.IsZero() {
		analysis.GeneratedAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		item.Analysis = &analysis
		item.UpdatedAt = time.Now().UTC()
		s.items[i] = item
		return item, nil
	}
	return Publication{}, ErrPublicationNotFound
}

func (s *PublicationStore) Validate(id string, validation PublicationValidation) (Publication, error) {
	id = strings.TrimSpace(id)
	validation.ValidatedBy = strings.TrimSpace(strings.ToLower(validation.ValidatedBy))
	validation.Notes = strings.TrimSpace(validation.Notes)
	if id == "" || validation.ValidatedBy == "" || validation.FinalDeadlineAt.IsZero() {
		return Publication{}, ErrInvalidPublication
	}
	if validation.ValidatedAt.IsZero() {
		validation.ValidatedAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		if item.Analysis == nil {
			return Publication{}, ErrPublicationNotReady
		}
		item.Validation = &validation
		item.UpdatedAt = time.Now().UTC()
		s.items[i] = item
		return item, nil
	}
	return Publication{}, ErrPublicationNotFound
}

func (s *PublicationStore) AttachTask(id, taskID string) (Publication, error) {
	id = strings.TrimSpace(id)
	taskID = strings.TrimSpace(taskID)
	if id == "" || taskID == "" {
		return Publication{}, ErrInvalidPublication
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		item.TaskID = taskID
		item.UpdatedAt = time.Now().UTC()
		s.items[i] = item
		return item, nil
	}
	return Publication{}, ErrPublicationNotFound
}

func (s *PublicationStore) PurgeOlderThan(cutoff time.Time) int {
	if cutoff.IsZero() {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	keep := s.items[:0]
	removed := 0
	for _, item := range s.items {
		if item.CreatedAt.Before(cutoff) {
			removed++
			continue
		}
		keep = append(keep, item)
	}
	s.items = keep
	return removed
}

func newPublicationID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "pub-fallback"
	}
	return "pub-" + hex.EncodeToString(buf)
}
