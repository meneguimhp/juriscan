package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"juriscan-backend/internal/identity/auth"
)

var (
	ErrInvalidUserPayload = errors.New("user: invalid payload")
	ErrUserNotFound       = errors.New("user: not found")
	ErrUserAlreadyExists  = errors.New("user: already exists")
	ErrUserInactive       = errors.New("user: inactive")
)

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

var allowedUserRoles = map[string]struct{}{
	"admin":      {},
	"controller": {},
	"lawyer":     {},
	"commercial": {},
}

type UserRecord struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type UserUpdate struct {
	Name   *string
	Role   *string
	Status *string
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) (*UserStore, error) {
	if db == nil {
		return nil, errors.New("user: nil db")
	}
	store := &UserStore{db: db}
	if err := store.ensureSchema(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *UserStore) ensureSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			role VARCHAR(32) NOT NULL,
			status VARCHAR(32) NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			last_login_at TEXT
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserStore) EnsureBootstrapUser(name, email, role string) error {
	_, err := s.Create(UserRecord{
		Name:   name,
		Email:  email,
		Role:   role,
		Status: UserStatusActive,
	})
	if errors.Is(err, ErrUserAlreadyExists) {
		return nil
	}
	return err
}

func (s *UserStore) List() []UserRecord {
	rows, err := s.db.Query(`
SELECT id, name, email, role, status, created_at, updated_at, last_login_at
FROM users
ORDER BY created_at ASC
`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]UserRecord, 0, 32)
	for rows.Next() {
		record, scanErr := scanUser(rows)
		if scanErr != nil {
			continue
		}
		out = append(out, record)
	}
	return out
}

func (s *UserStore) Create(input UserRecord) (UserRecord, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	role := strings.ToLower(strings.TrimSpace(input.Role))
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = UserStatusActive
	}
	if name == "" || email == "" || !isAllowedRole(role) || !isAllowedStatus(status) {
		return UserRecord{}, ErrInvalidUserPayload
	}

	now := time.Now().UTC()
	created := UserRecord{
		ID:        newUserID(),
		Name:      name,
		Email:     email,
		Role:      role,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.Exec(`
INSERT INTO users (id, name, email, role, status, created_at, updated_at, last_login_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULL)
`, created.ID, created.Name, created.Email, created.Role, created.Status, created.CreatedAt.Format(time.RFC3339Nano), created.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		if isUserEmailUniqueViolation(err) {
			return UserRecord{}, ErrUserAlreadyExists
		}
		return UserRecord{}, err
	}

	return created, nil
}

func (s *UserStore) Update(id string, update UserUpdate) (UserRecord, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return UserRecord{}, ErrInvalidUserPayload
	}

	item, err := s.getByID(id)
	if err != nil {
		return UserRecord{}, err
	}

	if update.Name != nil {
		name := strings.TrimSpace(*update.Name)
		if name == "" {
			return UserRecord{}, ErrInvalidUserPayload
		}
		item.Name = name
	}
	if update.Role != nil {
		role := strings.ToLower(strings.TrimSpace(*update.Role))
		if !isAllowedRole(role) {
			return UserRecord{}, ErrInvalidUserPayload
		}
		item.Role = role
	}
	if update.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*update.Status))
		if !isAllowedStatus(status) {
			return UserRecord{}, ErrInvalidUserPayload
		}
		item.Status = status
	}

	item.UpdatedAt = time.Now().UTC()
	_, err = s.db.Exec(`
UPDATE users
SET name = ?, role = ?, status = ?, updated_at = ?
WHERE id = ?
`, item.Name, item.Role, item.Status, item.UpdatedAt.Format(time.RFC3339Nano), item.ID)
	if err != nil {
		return UserRecord{}, err
	}

	return item, nil
}

func (s *UserStore) ResolveAuthUser(email string) (auth.User, error) {
	item, err := s.getByEmail(email)
	if err != nil {
		return auth.User{}, err
	}
	if item.Status != UserStatusActive {
		return auth.User{}, ErrUserInactive
	}
	return auth.User{
		ID:    item.ID,
		Email: item.Email,
		Role:  item.Role,
	}, nil
}

func (s *UserStore) MarkLogin(userID string, at time.Time) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	formatted := at.UTC().Format(time.RFC3339Nano)
	_, _ = s.db.Exec(`
UPDATE users
SET last_login_at = ?, updated_at = ?
WHERE id = ?
`, formatted, formatted, userID)
}

func (s *UserStore) getByID(id string) (UserRecord, error) {
	row := s.db.QueryRow(`
SELECT id, name, email, role, status, created_at, updated_at, last_login_at
FROM users
WHERE id = ?
`, strings.TrimSpace(id))
	record, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return UserRecord{}, ErrUserNotFound
	}
	if err != nil {
		return UserRecord{}, err
	}
	return record, nil
}

func (s *UserStore) getByEmail(email string) (UserRecord, error) {
	row := s.db.QueryRow(`
SELECT id, name, email, role, status, created_at, updated_at, last_login_at
FROM users
WHERE email = ?
`, strings.ToLower(strings.TrimSpace(email)))
	record, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return UserRecord{}, ErrUserNotFound
	}
	if err != nil {
		return UserRecord{}, err
	}
	return record, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (UserRecord, error) {
	var (
		record         UserRecord
		createdAtRaw   string
		updatedAtRaw   string
		lastLoginAtRaw sql.NullString
	)
	if err := scanner.Scan(
		&record.ID,
		&record.Name,
		&record.Email,
		&record.Role,
		&record.Status,
		&createdAtRaw,
		&updatedAtRaw,
		&lastLoginAtRaw,
	); err != nil {
		return UserRecord{}, err
	}

	record.Email = strings.ToLower(strings.TrimSpace(record.Email))
	record.Role = strings.ToLower(strings.TrimSpace(record.Role))
	record.Status = strings.ToLower(strings.TrimSpace(record.Status))

	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		return UserRecord{}, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtRaw)
	if err != nil {
		return UserRecord{}, err
	}
	record.CreatedAt = createdAt.UTC()
	record.UpdatedAt = updatedAt.UTC()

	if lastLoginAtRaw.Valid && strings.TrimSpace(lastLoginAtRaw.String) != "" {
		lastLoginAt, parseErr := time.Parse(time.RFC3339Nano, lastLoginAtRaw.String)
		if parseErr == nil {
			loginAt := lastLoginAt.UTC()
			record.LastLoginAt = &loginAt
		}
	}

	return record, nil
}

func isAllowedRole(role string) bool {
	_, ok := allowedUserRoles[strings.ToLower(strings.TrimSpace(role))]
	return ok
}

func isAllowedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case UserStatusActive, UserStatusInactive:
		return true
	default:
		return false
	}
}

func isUserEmailUniqueViolation(err error) bool {
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return (strings.Contains(msg, "unique") && strings.Contains(msg, "users.email")) ||
		strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "error 1062")
}

func newUserID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "user-fallback"
	}
	return "user-" + hex.EncodeToString(buf)
}

func deriveNameFromEmail(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	parts := strings.Split(email, "@")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "Admin"
	}
	local := strings.ReplaceAll(parts[0], ".", " ")
	local = strings.ReplaceAll(local, "_", " ")
	local = strings.TrimSpace(local)
	if local == "" {
		return "Admin"
	}
	return titleWords(local)
}

func titleWords(value string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		if runes[0] >= 'a' && runes[0] <= 'z' {
			runes[0] = runes[0] - 32
		}
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
