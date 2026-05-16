package app

import (
	"path/filepath"
	"testing"
	"time"
)

func TestUserStoreBootstrapAdmin(t *testing.T) {
	store := newTestUserStore(t)
	if err := store.EnsureBootstrapUser("Admin", "admin@juriscan.local", "admin"); err != nil {
		t.Fatalf("bootstrap user failed: %v", err)
	}
	items := store.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 bootstrap user, got %d", len(items))
	}
	if items[0].Role != "admin" {
		t.Fatalf("expected admin role, got %q", items[0].Role)
	}
	if items[0].Status != UserStatusActive {
		t.Fatalf("expected active status, got %q", items[0].Status)
	}
}

func TestUserStoreCreateResolveAndMarkLogin(t *testing.T) {
	store := newTestUserStore(t)
	created, err := store.Create(UserRecord{
		Name:   "Comercial One",
		Email:  "commercial@juriscan.local",
		Role:   "commercial",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	resolved, err := store.ResolveAuthUser("commercial@juriscan.local")
	if err != nil {
		t.Fatalf("resolve auth user failed: %v", err)
	}
	if resolved.ID != created.ID || resolved.Role != "commercial" {
		t.Fatalf("unexpected resolved user: %+v", resolved)
	}

	loginAt := time.Now().UTC()
	store.MarkLogin(created.ID, loginAt)
	items := store.List()
	if items[0].LastLoginAt == nil {
		t.Fatal("expected last login to be set")
	}
}

func TestUserStoreBlocksInactiveUser(t *testing.T) {
	store := newTestUserStore(t)
	created, err := store.Create(UserRecord{
		Name:   "Controller One",
		Email:  "controller@juriscan.local",
		Role:   "controller",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	inactive := UserStatusInactive
	if _, err := store.Update(created.ID, UserUpdate{Status: &inactive}); err != nil {
		t.Fatalf("update user status failed: %v", err)
	}

	if _, err := store.ResolveAuthUser("controller@juriscan.local"); err == nil {
		t.Fatal("expected inactive user to be blocked")
	}
}

func newTestUserStore(t *testing.T) *UserStore {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "users.db")
	db, err := openSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	store, err := NewUserStore(db)
	if err != nil {
		t.Fatalf("new user store failed: %v", err)
	}
	return store
}
