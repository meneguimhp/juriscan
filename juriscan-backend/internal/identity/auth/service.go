package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInvalidEmail       = errors.New("auth: invalid email")
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type otpToken struct {
	Code      string
	ExpiresAt time.Time
}

type session struct {
	User      User
	ExpiresAt time.Time
}

type Service struct {
	nowFn func() time.Time

	ttlOTP     time.Duration
	ttlSession time.Duration

	resolveUserFn func(email string) (User, error)

	mu       sync.RWMutex
	otpByKey map[string]otpToken
	sessions map[string]session
}

func NewService(ttlOTP, ttlSession time.Duration, resolveUserFn func(email string) (User, error)) *Service {
	if ttlOTP <= 0 {
		ttlOTP = 10 * time.Minute
	}
	if ttlSession <= 0 {
		ttlSession = 24 * time.Hour
	}
	if resolveUserFn == nil {
		resolveUserFn = func(email string) (User, error) {
			email = strings.ToLower(strings.TrimSpace(email))
			return User{
				ID:    makeUserID(email),
				Email: email,
				Role:  "commercial",
			}, nil
		}
	}
	return &Service{
		nowFn:         time.Now,
		ttlOTP:        ttlOTP,
		ttlSession:    ttlSession,
		resolveUserFn: resolveUserFn,
		otpByKey:      make(map[string]otpToken),
		sessions:      make(map[string]session),
	}
}

func (s *Service) RequestOTP(email string) (code string, expiresInSec int, err error) {
	email, err = normalizeEmail(email)
	if err != nil {
		return "", 0, err
	}
	if _, err := s.resolveUserFn(email); err != nil {
		return "", 0, ErrInvalidCredentials
	}

	code, err = generateNumericCode(6)
	if err != nil {
		return "", 0, err
	}

	s.mu.Lock()
	s.otpByKey[email] = otpToken{
		Code:      code,
		ExpiresAt: s.nowFn().Add(s.ttlOTP),
	}
	s.mu.Unlock()

	return code, int(s.ttlOTP.Seconds()), nil
}

func (s *Service) VerifyOTP(email, code string) (string, User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return "", User{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.otpByKey[email]
	if !ok || token.ExpiresAt.Before(s.nowFn()) || strings.TrimSpace(code) != token.Code {
		return "", User{}, ErrInvalidCredentials
	}
	delete(s.otpByKey, email)

	user, err := s.resolveUserFn(email)
	if err != nil {
		return "", User{}, ErrInvalidCredentials
	}
	user.Email = email
	if strings.TrimSpace(user.ID) == "" {
		user.ID = makeUserID(email)
	}

	sessionToken, err := generateHexToken(32)
	if err != nil {
		return "", User{}, err
	}

	s.sessions[sessionToken] = session{
		User:      user,
		ExpiresAt: s.nowFn().Add(s.ttlSession),
	}

	return sessionToken, user, nil
}

func (s *Service) ValidateSession(token string) (User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, ErrInvalidCredentials
	}

	s.mu.RLock()
	current, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok || current.ExpiresAt.Before(s.nowFn()) {
		return User{}, ErrInvalidCredentials
	}

	return current.User, nil
}

func (s *Service) RevokeSession(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func generateNumericCode(size int) (string, error) {
	if size <= 0 {
		size = 6
	}
	b := strings.Builder{}
	b.Grow(size)
	for i := 0; i < size; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteString(fmt.Sprintf("%d", n.Int64()))
	}
	return b.String(), nil
}

func generateHexToken(size int) (string, error) {
	if size <= 0 {
		size = 32
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func makeUserID(email string) string {
	raw := []byte(strings.ToLower(strings.TrimSpace(email)))
	if len(raw) == 0 {
		return "user-unknown"
	}
	if len(raw) > 8 {
		raw = raw[:8]
	}
	return "user-" + hex.EncodeToString(raw)
}
