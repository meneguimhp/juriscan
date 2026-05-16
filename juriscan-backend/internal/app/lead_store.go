package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidLead  = errors.New("lead: invalid payload")
	ErrLeadNotFound = errors.New("lead: not found")
)

type LeadStore struct {
	mu    sync.RWMutex
	items []Lead
}

type LeadUpdate struct {
	Name       *string
	Phone      *string
	Origin     *string
	Subject    *string
	OwnerEmail *string
	Stage      *string
	NextStep   *string
}

var allowedLeadStages = map[string]struct{}{
	"novo":        {},
	"triado":      {},
	"qualificado": {},
	"proposta":    {},
	"fechado":     {},
	"perdido":     {},
}

func NewLeadStore() *LeadStore {
	return &LeadStore{
		items: make([]Lead, 0, 32),
	}
}

func (s *LeadStore) List() []Lead {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Lead, len(s.items))
	for i := range s.items {
		out[i] = s.items[i]
		if s.items[i].FirstResponseAt != nil {
			firstResponse := *s.items[i].FirstResponseAt
			out[i].FirstResponseAt = &firstResponse
		}
		if len(s.items[i].History) > 0 {
			out[i].History = append([]LeadHistoryEvent(nil), s.items[i].History...)
		}
		if s.items[i].AIClassification != nil {
			copyClass := *s.items[i].AIClassification
			if s.items[i].AIClassification.OverriddenAt != nil {
				overriddenAt := *s.items[i].AIClassification.OverriddenAt
				copyClass.OverriddenAt = &overriddenAt
			}
			out[i].AIClassification = &copyClass
		}
		if s.items[i].NextFollowUpAt != nil {
			nextFollowUp := *s.items[i].NextFollowUpAt
			out[i].NextFollowUpAt = &nextFollowUp
		}
	}
	return out
}

func (s *LeadStore) GetByID(id string) (Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.items {
		if s.items[i].ID != strings.TrimSpace(id) {
			continue
		}
		item := s.items[i]
		if item.FirstResponseAt != nil {
			firstResponse := *item.FirstResponseAt
			item.FirstResponseAt = &firstResponse
		}
		if item.AIClassification != nil {
			copyClass := *item.AIClassification
			if item.AIClassification.OverriddenAt != nil {
				overriddenAt := *item.AIClassification.OverriddenAt
				copyClass.OverriddenAt = &overriddenAt
			}
			item.AIClassification = &copyClass
		}
		if item.NextFollowUpAt != nil {
			nextFollowUp := *item.NextFollowUpAt
			item.NextFollowUpAt = &nextFollowUp
		}
		if len(item.History) > 0 {
			item.History = append([]LeadHistoryEvent(nil), item.History...)
		}
		return item, nil
	}
	return Lead{}, ErrLeadNotFound
}

func (s *LeadStore) FindByPhone(phone string) []Lead {
	normalized := normalizePhone(phone)
	if normalized == "" {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Lead, 0, 4)
	for i := range s.items {
		if normalizePhone(s.items[i].Phone) != normalized {
			continue
		}
		out = append(out, s.items[i])
	}
	return out
}

func (s *LeadStore) Create(input Lead) (Lead, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Stage = strings.TrimSpace(input.Stage)
	if input.Name == "" || input.Phone == "" {
		return Lead{}, ErrInvalidLead
	}

	if input.Stage == "" {
		input.Stage = "novo"
	}
	if !isValidLeadStage(input.Stage) {
		return Lead{}, ErrInvalidLead
	}
	now := time.Now().UTC()
	input.ID = newLeadID()
	input.CreatedAt = now
	input.UpdatedAt = now
	input.History = []LeadHistoryEvent{
		{
			Type: "created",
			At:   now,
			Note: "lead criado",
		},
	}

	s.mu.Lock()
	s.items = append(s.items, input)
	s.mu.Unlock()

	return input, nil
}

func (s *LeadStore) UpdateStage(id, stage string) (Lead, error) {
	stage = strings.TrimSpace(strings.ToLower(stage))
	if !isValidLeadStage(stage) {
		return Lead{}, ErrInvalidLead
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		prev := s.items[i].Stage
		s.items[i].Stage = stage
		s.items[i].UpdatedAt = time.Now().UTC()
		if prev == "novo" && stage != "novo" && s.items[i].FirstResponseAt == nil {
			firstResponse := s.items[i].UpdatedAt
			s.items[i].FirstResponseAt = &firstResponse
		}
		s.items[i].History = append(s.items[i].History, LeadHistoryEvent{
			Type:      "stage_changed",
			At:        s.items[i].UpdatedAt,
			FromStage: prev,
			ToStage:   stage,
			Note:      "etapa alterada",
		})
		return s.items[i], nil
	}
	return Lead{}, ErrLeadNotFound
}

func (s *LeadStore) Update(id string, update LeadUpdate) (Lead, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}

		item := s.items[i]
		prevStage := item.Stage
		if update.Name != nil {
			item.Name = strings.TrimSpace(*update.Name)
			if item.Name == "" {
				return Lead{}, ErrInvalidLead
			}
		}
		if update.Phone != nil {
			item.Phone = strings.TrimSpace(*update.Phone)
			if item.Phone == "" {
				return Lead{}, ErrInvalidLead
			}
		}
		if update.Origin != nil {
			item.Origin = strings.TrimSpace(*update.Origin)
		}
		if update.Subject != nil {
			item.Subject = strings.TrimSpace(*update.Subject)
		}
		if update.OwnerEmail != nil {
			item.OwnerEmail = strings.TrimSpace(*update.OwnerEmail)
		}
		if update.Stage != nil {
			stage := strings.TrimSpace(strings.ToLower(*update.Stage))
			if !isValidLeadStage(stage) {
				return Lead{}, ErrInvalidLead
			}
			item.Stage = stage
		}
		if update.NextStep != nil {
			item.NextStep = strings.TrimSpace(*update.NextStep)
		}
		item.UpdatedAt = time.Now().UTC()
		if prevStage == "novo" && item.Stage != "novo" && item.FirstResponseAt == nil {
			firstResponse := item.UpdatedAt
			item.FirstResponseAt = &firstResponse
		}
		event := LeadHistoryEvent{
			Type: "updated",
			At:   item.UpdatedAt,
			Note: "dados do lead atualizados",
		}
		if prevStage != item.Stage {
			event.FromStage = prevStage
			event.ToStage = item.Stage
		}
		item.History = append(item.History, event)

		s.items[i] = item
		return item, nil
	}

	return Lead{}, ErrLeadNotFound
}

func (s *LeadStore) ApplyAIClassification(id string, classification LeadAIClassification) (Lead, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Lead{}, ErrInvalidLead
	}
	classification.Category = strings.TrimSpace(strings.ToLower(classification.Category))
	classification.Urgency = strings.TrimSpace(strings.ToLower(classification.Urgency))
	classification.SuggestedAction = strings.TrimSpace(classification.SuggestedAction)
	classification.Justification = strings.TrimSpace(classification.Justification)
	classification.Model = strings.TrimSpace(classification.Model)
	if classification.Category == "" || classification.Urgency == "" || classification.Score < 0 || classification.Score > 100 {
		return Lead{}, ErrInvalidLead
	}
	if classification.GeneratedAt.IsZero() {
		classification.GeneratedAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		item.AIClassification = &classification
		if classification.SuggestedAction != "" {
			item.NextStep = classification.SuggestedAction
		}
		item.UpdatedAt = time.Now().UTC()
		item.History = append(item.History, LeadHistoryEvent{
			Type: "ai_triaged",
			At:   item.UpdatedAt,
			Note: "triagem automatica aplicada",
		})
		s.items[i] = item
		return item, nil
	}
	return Lead{}, ErrLeadNotFound
}

func (s *LeadStore) OverrideAIClassification(id, reason, by string, override LeadAIClassification) (Lead, error) {
	id = strings.TrimSpace(id)
	reason = strings.TrimSpace(reason)
	by = strings.TrimSpace(strings.ToLower(by))
	if id == "" || reason == "" || by == "" {
		return Lead{}, ErrInvalidLead
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		if item.AIClassification == nil {
			return Lead{}, ErrInvalidLead
		}
		override.Category = strings.TrimSpace(strings.ToLower(override.Category))
		override.Urgency = strings.TrimSpace(strings.ToLower(override.Urgency))
		override.SuggestedAction = strings.TrimSpace(override.SuggestedAction)
		override.Justification = strings.TrimSpace(override.Justification)
		override.Model = strings.TrimSpace(override.Model)
		if override.Category == "" || override.Urgency == "" || override.Score < 0 || override.Score > 100 {
			return Lead{}, ErrInvalidLead
		}
		now := time.Now().UTC()
		override.GeneratedAt = now
		override.OverriddenBy = by
		override.OverriddenAt = &now
		override.OverrideReason = reason
		item.AIClassification = &override
		if override.SuggestedAction != "" {
			item.NextStep = override.SuggestedAction
		}
		item.UpdatedAt = now
		item.History = append(item.History, LeadHistoryEvent{
			Type: "ai_override",
			At:   now,
			Note: reason,
		})
		s.items[i] = item
		return item, nil
	}
	return Lead{}, ErrLeadNotFound
}

func (s *LeadStore) SetNextFollowUp(id string, at time.Time) (Lead, error) {
	id = strings.TrimSpace(id)
	if id == "" || at.IsZero() {
		return Lead{}, ErrInvalidLead
	}
	utcAt := at.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID != id {
			continue
		}
		item := s.items[i]
		item.NextFollowUpAt = &utcAt
		item.UpdatedAt = time.Now().UTC()
		item.History = append(item.History, LeadHistoryEvent{
			Type: "followup_scheduled",
			At:   item.UpdatedAt,
			Note: utcAt.Format(time.RFC3339),
		})
		s.items[i] = item
		return item, nil
	}
	return Lead{}, ErrLeadNotFound
}

func newLeadID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "lead-fallback"
	}
	return "lead-" + hex.EncodeToString(buf)
}

func isValidLeadStage(stage string) bool {
	_, ok := allowedLeadStages[strings.ToLower(strings.TrimSpace(stage))]
	return ok
}
