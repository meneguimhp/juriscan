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
	ErrInvalidConversation  = errors.New("whatsapp: invalid payload")
	ErrConversationNotFound = errors.New("whatsapp: conversation not found")
)

const (
	ConversationStatusNew    = "nova"
	ConversationStatusNoLead = "sem_lead"
	ConversationStatusLinked = "vinculada"
)

type ConversationStore struct {
	mu        sync.RWMutex
	byID      map[string]WhatsAppConversation
	idByPhone map[string]string
}

func NewConversationStore() *ConversationStore {
	return &ConversationStore{
		byID:      make(map[string]WhatsAppConversation, 32),
		idByPhone: make(map[string]string, 32),
	}
}

func (s *ConversationStore) List() []WhatsAppConversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]WhatsAppConversation, 0, len(s.byID))
	for _, item := range s.byID {
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *ConversationStore) Ingest(input WhatsAppInboundMessage) (WhatsAppConversation, error) {
	phone := normalizePhone(input.Phone)
	message := strings.TrimSpace(input.Message)
	if len(phone) < 10 || message == "" {
		return WhatsAppConversation{}, ErrInvalidConversation
	}

	now := time.Now().UTC()
	receivedAt := input.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = now
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.idByPhone[phone]; ok {
		item := s.byID[id]
		item.LastMessage = message
		item.LastMessageAt = receivedAt
		item.UpdatedAt = now
		item.MessageCount++
		if strings.TrimSpace(input.ContactName) != "" {
			item.ContactName = strings.TrimSpace(input.ContactName)
		}
		s.byID[id] = item
		return item, nil
	}

	id := strings.TrimSpace(input.ExternalID)
	if id == "" {
		id = newConversationID()
	}
	item := WhatsAppConversation{
		ID:            id,
		Phone:         phone,
		ContactName:   strings.TrimSpace(input.ContactName),
		LastMessage:   message,
		Status:        ConversationStatusNew,
		MessageCount:  1,
		LastMessageAt: receivedAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.byID[id] = item
	s.idByPhone[phone] = id

	return item, nil
}

func (s *ConversationStore) UpdateClassification(id, status, leadID string) (WhatsAppConversation, error) {
	id = strings.TrimSpace(id)
	status = strings.TrimSpace(strings.ToLower(status))
	leadID = strings.TrimSpace(leadID)

	switch status {
	case ConversationStatusNew, ConversationStatusNoLead:
		leadID = ""
	case ConversationStatusLinked:
		if leadID == "" {
			return WhatsAppConversation{}, ErrInvalidConversation
		}
	default:
		return WhatsAppConversation{}, ErrInvalidConversation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.byID[id]
	if !ok {
		return WhatsAppConversation{}, ErrConversationNotFound
	}

	item.Status = status
	item.LeadID = leadID
	item.UpdatedAt = time.Now().UTC()
	s.byID[id] = item

	return item, nil
}

func normalizePhone(raw string) string {
	digits := strings.Builder{}
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	return digits.String()
}

func newConversationID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "conv-fallback"
	}
	return "conv-" + hex.EncodeToString(buf)
}
