package app

import (
	"testing"
	"time"
)

func TestBuildLeadQueueMarksOverdueLead(t *testing.T) {
	now := time.Date(2026, 4, 12, 22, 0, 0, 0, time.UTC)
	createdAt := now.Add(-45 * time.Minute)
	items := []Lead{
		{
			ID:        "lead-1",
			Name:      "Lead A",
			Phone:     "11999999999",
			Stage:     "novo",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	queue := BuildLeadQueue(items, now, 30*time.Minute, false)
	if len(queue) != 1 {
		t.Fatalf("expected 1 item, got %d", len(queue))
	}
	if queue[0].SLAStatus != LeadSLAStatusOverdue {
		t.Fatalf("expected status %q, got %q", LeadSLAStatusOverdue, queue[0].SLAStatus)
	}
	if queue[0].MinutesWithoutResponse < 45 {
		t.Fatalf("expected at least 45 minutes, got %d", queue[0].MinutesWithoutResponse)
	}
}

func TestBuildLeadQueueFiltersOnlyOverdue(t *testing.T) {
	now := time.Date(2026, 4, 12, 22, 0, 0, 0, time.UTC)
	createdLate := now.Add(-45 * time.Minute)
	createdInTime := now.Add(-10 * time.Minute)

	queue := BuildLeadQueue([]Lead{
		{
			ID:        "lead-late",
			Name:      "Lead Late",
			Phone:     "11999999999",
			Stage:     "novo",
			CreatedAt: createdLate,
			UpdatedAt: createdLate,
		},
		{
			ID:        "lead-open",
			Name:      "Lead Open",
			Phone:     "11988887777",
			Stage:     "novo",
			CreatedAt: createdInTime,
			UpdatedAt: createdInTime,
		},
	}, now, 30*time.Minute, true)

	if len(queue) != 1 {
		t.Fatalf("expected 1 overdue item, got %d", len(queue))
	}
	if queue[0].ID != "lead-late" {
		t.Fatalf("expected lead-late, got %s", queue[0].ID)
	}
}
