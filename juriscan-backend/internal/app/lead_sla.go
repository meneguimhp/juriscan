package app

import (
	"sort"
	"time"
)

const (
	LeadSLAStatusOpenInTime   = "em_aberto"
	LeadSLAStatusOverdue      = "estourado"
	LeadSLAStatusServedInTime = "atendido_no_prazo"
	LeadSLAStatusServedLate   = "atendido_fora_prazo"
)

func BuildLeadQueue(items []Lead, now time.Time, sla time.Duration, onlyOverdue bool) []LeadQueueItem {
	if sla <= 0 {
		sla = 30 * time.Minute
	}

	out := make([]LeadQueueItem, 0, len(items))
	for _, item := range items {
		queueItem := LeadQueueItem{Lead: item}
		dueAt := item.CreatedAt.Add(sla)
		queueItem.ResponseDueAt = &dueAt

		if item.FirstResponseAt == nil {
			queueItem.MinutesWithoutResponse = maxInt64(0, int64(now.Sub(item.CreatedAt).Minutes()))
			if now.After(dueAt) {
				queueItem.SLAStatus = LeadSLAStatusOverdue
			} else {
				queueItem.SLAStatus = LeadSLAStatusOpenInTime
			}
		} else {
			queueItem.MinutesWithoutResponse = maxInt64(0, int64(item.FirstResponseAt.Sub(item.CreatedAt).Minutes()))
			if item.FirstResponseAt.After(dueAt) {
				queueItem.SLAStatus = LeadSLAStatusServedLate
			} else {
				queueItem.SLAStatus = LeadSLAStatusServedInTime
			}
		}

		if onlyOverdue && queueItem.SLAStatus != LeadSLAStatusOverdue {
			continue
		}
		out = append(out, queueItem)
	}

	sort.SliceStable(out, func(i, j int) bool {
		leftOverdue := out[i].SLAStatus == LeadSLAStatusOverdue
		rightOverdue := out[j].SLAStatus == LeadSLAStatusOverdue
		if leftOverdue != rightOverdue {
			return leftOverdue
		}
		if out[i].MinutesWithoutResponse != out[j].MinutesWithoutResponse {
			return out[i].MinutesWithoutResponse > out[j].MinutesWithoutResponse
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
