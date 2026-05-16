package app

import "testing"

func TestLeadStoreCreateAndUpdateStage(t *testing.T) {
	store := NewLeadStore()
	lead, err := store.Create(Lead{
		Name:  "Cliente Teste",
		Phone: "+5511999999999",
	})
	if err != nil {
		t.Fatalf("create lead failed: %v", err)
	}
	if lead.Stage != "novo" {
		t.Fatalf("expected stage novo, got %q", lead.Stage)
	}

	updated, err := store.UpdateStage(lead.ID, "qualificado")
	if err != nil {
		t.Fatalf("update stage failed: %v", err)
	}
	if updated.Stage != "qualificado" {
		t.Fatalf("expected qualificado, got %q", updated.Stage)
	}
}

func TestLeadStoreRejectsInvalidStage(t *testing.T) {
	store := NewLeadStore()
	lead, err := store.Create(Lead{
		Name:  "Cliente Teste",
		Phone: "+5511999999999",
	})
	if err != nil {
		t.Fatalf("create lead failed: %v", err)
	}

	if _, err := store.UpdateStage(lead.ID, "inexistente"); err == nil {
		t.Fatal("expected invalid stage error")
	}
}

func TestLeadStoreUpdateFields(t *testing.T) {
	store := NewLeadStore()
	lead, err := store.Create(Lead{
		Name:       "Cliente A",
		Phone:      "11988887777",
		Origin:     "whatsapp",
		Subject:    "Dúvida inicial",
		OwnerEmail: "admin@juriscan.local",
	})
	if err != nil {
		t.Fatalf("create lead failed: %v", err)
	}

	name := "Cliente Atualizado"
	phone := "11977776666"
	subject := "Reunião agendada"
	stage := "qualificado"
	updated, err := store.Update(lead.ID, LeadUpdate{
		Name:    &name,
		Phone:   &phone,
		Subject: &subject,
		Stage:   &stage,
	})
	if err != nil {
		t.Fatalf("update lead failed: %v", err)
	}

	if updated.Name != name || updated.Phone != phone || updated.Subject != subject || updated.Stage != stage {
		t.Fatalf("unexpected updated lead: %+v", updated)
	}
}

func TestLeadStoreUpdateRejectsInvalidPayload(t *testing.T) {
	store := NewLeadStore()
	lead, err := store.Create(Lead{
		Name:  "Cliente Teste",
		Phone: "11999999999",
	})
	if err != nil {
		t.Fatalf("create lead failed: %v", err)
	}

	invalidName := "   "
	if _, err := store.Update(lead.ID, LeadUpdate{Name: &invalidName}); err == nil {
		t.Fatal("expected invalid payload error")
	}
}
