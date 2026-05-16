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
	ErrTemplateNotFound = errors.New("template: not found")
	ErrInvalidTemplate  = errors.New("template: invalid payload")
	ErrFollowUpNotFound = errors.New("followup: not found")
	ErrInvalidFollowUp  = errors.New("followup: invalid payload")
)

const (
	FollowUpStatusPending   = "pendente"
	FollowUpStatusDone      = "concluido"
	FollowUpStatusCancelled = "cancelado"
)

type TemplateFollowUpStore struct {
	mu          sync.RWMutex
	templates   []MessageTemplate
	followUps   []FollowUp
	templateMap map[string]MessageTemplate
}

func NewTemplateFollowUpStore() *TemplateFollowUpStore {
	now := time.Now().UTC()
	defaultTemplates := []MessageTemplate{
		{
			ID:        newTemplateID(),
			Name:      "Primeiro contato",
			Channel:   "whatsapp",
			Body:      "Ola! Recebemos seu contato e vamos analisar seu caso.",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        newTemplateID(),
			Name:      "Solicitacao de documentos",
			Channel:   "whatsapp",
			Body:      "Para avancarmos, por favor envie os documentos principais do caso.",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        newTemplateID(),
			Name:      "Agendamento de retorno",
			Channel:   "whatsapp",
			Body:      "Podemos agendar uma conversa rapida para alinhar os proximos passos?",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	templateMap := make(map[string]MessageTemplate, len(defaultTemplates))
	for _, tpl := range defaultTemplates {
		templateMap[tpl.ID] = tpl
	}
	return &TemplateFollowUpStore{
		templates:   defaultTemplates,
		followUps:   make([]FollowUp, 0, 64),
		templateMap: templateMap,
	}
}

func (s *TemplateFollowUpStore) ListTemplates() []MessageTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := append([]MessageTemplate(nil), s.templates...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *TemplateFollowUpStore) CreateTemplate(input MessageTemplate) (MessageTemplate, error) {
	name := strings.TrimSpace(input.Name)
	channel := strings.TrimSpace(strings.ToLower(input.Channel))
	body := strings.TrimSpace(input.Body)
	if channel == "" {
		channel = "whatsapp"
	}
	if name == "" || body == "" {
		return MessageTemplate{}, ErrInvalidTemplate
	}

	now := time.Now().UTC()
	item := MessageTemplate{
		ID:        newTemplateID(),
		Name:      name,
		Channel:   channel,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	s.templates = append(s.templates, item)
	s.templateMap[item.ID] = item
	s.mu.Unlock()

	return item, nil
}

func (s *TemplateFollowUpStore) ResolveTemplate(templateID string) (MessageTemplate, error) {
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return MessageTemplate{}, ErrTemplateNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.templateMap[templateID]
	if !ok {
		return MessageTemplate{}, ErrTemplateNotFound
	}
	return item, nil
}

func (s *TemplateFollowUpStore) CreateFollowUp(input FollowUp) (FollowUp, error) {
	leadID := strings.TrimSpace(input.LeadID)
	message := strings.TrimSpace(input.Message)
	templateID := strings.TrimSpace(input.TemplateID)
	createdBy := strings.TrimSpace(strings.ToLower(input.CreatedBy))
	if leadID == "" || message == "" || createdBy == "" || input.DueAt.IsZero() {
		return FollowUp{}, ErrInvalidFollowUp
	}

	now := time.Now().UTC()
	item := FollowUp{
		ID:         newFollowUpID(),
		LeadID:     leadID,
		TemplateID: templateID,
		Message:    message,
		DueAt:      input.DueAt.UTC(),
		Status:     FollowUpStatusPending,
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	s.mu.Lock()
	s.followUps = append(s.followUps, item)
	s.mu.Unlock()

	return item, nil
}

func (s *TemplateFollowUpStore) ListFollowUps(leadID string, onlyPending bool) []FollowUp {
	leadID = strings.TrimSpace(leadID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]FollowUp, 0, len(s.followUps))
	for _, item := range s.followUps {
		if leadID != "" && item.LeadID != leadID {
			continue
		}
		if onlyPending && item.Status != FollowUpStatusPending {
			continue
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].DueAt.Before(out[j].DueAt)
	})
	return out
}

func (s *TemplateFollowUpStore) UpdateFollowUpStatus(id, status string) (FollowUp, error) {
	id = strings.TrimSpace(id)
	status = strings.TrimSpace(strings.ToLower(status))
	if id == "" {
		return FollowUp{}, ErrInvalidFollowUp
	}
	switch status {
	case FollowUpStatusPending, FollowUpStatusDone, FollowUpStatusCancelled:
	default:
		return FollowUp{}, ErrInvalidFollowUp
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.followUps {
		if s.followUps[i].ID != id {
			continue
		}
		item := s.followUps[i]
		item.Status = status
		item.UpdatedAt = time.Now().UTC()
		if status == FollowUpStatusDone {
			doneAt := item.UpdatedAt
			item.DoneAt = &doneAt
		} else {
			item.DoneAt = nil
		}
		s.followUps[i] = item
		return item, nil
	}
	return FollowUp{}, ErrFollowUpNotFound
}

func newTemplateID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "tpl-fallback"
	}
	return "tpl-" + hex.EncodeToString(buf)
}

func newFollowUpID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "fu-fallback"
	}
	return "fu-" + hex.EncodeToString(buf)
}
