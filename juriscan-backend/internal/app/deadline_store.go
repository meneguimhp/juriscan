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
	ErrDeadlineTaskNotFound = errors.New("deadline: task not found")
	ErrInvalidDeadlineTask  = errors.New("deadline: invalid payload")
)

const (
	DeadlineTaskStatusOpen      = "aberto"
	DeadlineTaskStatusInProgress = "em_execucao"
	DeadlineTaskStatusDone      = "concluido"
)

type DeadlineStore struct {
	mu    sync.RWMutex
	tasks []DeadlineTask
}

func NewDeadlineStore() *DeadlineStore {
	return &DeadlineStore{
		tasks: make([]DeadlineTask, 0, 64),
	}
}

func (s *DeadlineStore) ListTasks(ownerEmail, status string) []DeadlineTask {
	ownerEmail = strings.TrimSpace(strings.ToLower(ownerEmail))
	status = strings.TrimSpace(strings.ToLower(status))

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]DeadlineTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		if ownerEmail != "" && strings.ToLower(task.OwnerEmail) != ownerEmail {
			continue
		}
		if status != "" && strings.ToLower(task.Status) != status {
			continue
		}
		out = append(out, task)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].DueAt.Before(out[j].DueAt)
	})
	return out
}

func (s *DeadlineStore) CreateTask(input DeadlineTask) (DeadlineTask, error) {
	title := strings.TrimSpace(input.Title)
	publicationID := strings.TrimSpace(input.PublicationID)
	ownerEmail := strings.TrimSpace(strings.ToLower(input.OwnerEmail))
	risk := strings.TrimSpace(strings.ToLower(input.Risk))
	if title == "" || publicationID == "" || ownerEmail == "" || input.DueAt.IsZero() {
		return DeadlineTask{}, ErrInvalidDeadlineTask
	}
	if risk == "" {
		risk = "medio"
	}

	now := time.Now().UTC()
	item := DeadlineTask{
		ID:            newDeadlineTaskID(),
		PublicationID: publicationID,
		Title:         title,
		OwnerEmail:    ownerEmail,
		DueAt:         input.DueAt.UTC(),
		Status:        DeadlineTaskStatusOpen,
		Risk:          risk,
		Checklist:     append([]string(nil), input.Checklist...),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	s.mu.Lock()
	s.tasks = append(s.tasks, item)
	s.mu.Unlock()
	return item, nil
}

func (s *DeadlineStore) UpdateTask(id string, status, ownerEmail string) (DeadlineTask, error) {
	id = strings.TrimSpace(id)
	status = strings.TrimSpace(strings.ToLower(status))
	ownerEmail = strings.TrimSpace(strings.ToLower(ownerEmail))
	if id == "" {
		return DeadlineTask{}, ErrInvalidDeadlineTask
	}
	if status != "" {
		switch status {
		case DeadlineTaskStatusOpen, DeadlineTaskStatusInProgress, DeadlineTaskStatusDone:
		default:
			return DeadlineTask{}, ErrInvalidDeadlineTask
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID != id {
			continue
		}
		item := s.tasks[i]
		if status != "" {
			item.Status = status
		}
		if ownerEmail != "" {
			item.OwnerEmail = ownerEmail
		}
		item.UpdatedAt = time.Now().UTC()
		s.tasks[i] = item
		return item, nil
	}
	return DeadlineTask{}, ErrDeadlineTaskNotFound
}

func (s *DeadlineStore) BuildAlerts(now time.Time) []DeadlineAlert {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]DeadlineAlert, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Status == DeadlineTaskStatusDone {
			continue
		}
		diff := task.DueAt.Sub(now)
		switch {
		case diff < 0:
			alerts = append(alerts, DeadlineAlert{
				ID:        "alert-" + task.ID + "-overdue",
				TaskID:    task.ID,
				Type:      "atrasado",
				Message:   "Prazo atrasado",
				CreatedAt: now,
			})
		case diff <= 24*time.Hour:
			alerts = append(alerts, DeadlineAlert{
				ID:        "alert-" + task.ID + "-d0",
				TaskID:    task.ID,
				Type:      "d0",
				Message:   "Prazo vence hoje",
				CreatedAt: now,
			})
		case diff <= 48*time.Hour:
			alerts = append(alerts, DeadlineAlert{
				ID:        "alert-" + task.ID + "-d1",
				TaskID:    task.ID,
				Type:      "d1",
				Message:   "Prazo vence em ate 1 dia",
				CreatedAt: now,
			})
		}
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		return alerts[i].Type < alerts[j].Type
	})
	return alerts
}

func newDeadlineTaskID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "task-fallback"
	}
	return "task-" + hex.EncodeToString(buf)
}
