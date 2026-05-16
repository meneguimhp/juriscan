package httpx

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]any{
		"ok": true,
	}

	WriteJSON(rec, 201, payload)

	if rec.Code != 201 {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", got)
	}

	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if decoded["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", decoded["ok"])
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, 400, "payload invalido")

	if rec.Code != 400 {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if decoded["error"] != "payload invalido" {
		t.Fatalf("expected error message, got %#v", decoded["error"])
	}
}
