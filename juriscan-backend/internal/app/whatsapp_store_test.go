package app

import "testing"

func TestConversationStoreIngestAndClassify(t *testing.T) {
	store := NewConversationStore()

	created, err := store.Ingest(WhatsAppInboundMessage{
		Phone:       "(11) 99999-0000",
		Message:     "Oi, quero atendimento",
		ContactName: "Cliente Teste",
	})
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if created.Status != ConversationStatusNew {
		t.Fatalf("expected status %q, got %q", ConversationStatusNew, created.Status)
	}
	if created.MessageCount != 1 {
		t.Fatalf("expected message count 1, got %d", created.MessageCount)
	}

	updated, err := store.UpdateClassification(created.ID, ConversationStatusLinked, "lead-123")
	if err != nil {
		t.Fatalf("update classification failed: %v", err)
	}
	if updated.Status != ConversationStatusLinked {
		t.Fatalf("expected linked status, got %q", updated.Status)
	}
	if updated.LeadID != "lead-123" {
		t.Fatalf("expected linked lead id, got %q", updated.LeadID)
	}
}

func TestConversationStoreIngestMergesByPhone(t *testing.T) {
	store := NewConversationStore()

	first, err := store.Ingest(WhatsAppInboundMessage{
		Phone:   "11911112222",
		Message: "Primeira mensagem",
	})
	if err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}

	second, err := store.Ingest(WhatsAppInboundMessage{
		Phone:   "11 91111-2222",
		Message: "Segunda mensagem",
	})
	if err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same conversation id, got %q and %q", first.ID, second.ID)
	}
	if second.MessageCount != 2 {
		t.Fatalf("expected message count 2, got %d", second.MessageCount)
	}
}
