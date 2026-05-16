package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"juriscan-backend/internal/httpx"
	"juriscan-backend/internal/identity/auth"
)

type ctxKey string

const userCtxKey ctxKey = "principal_user"
const sessionTokenCtxKey ctxKey = "session_token"

func userFromContext(ctx context.Context) (auth.User, bool) {
	v := ctx.Value(userCtxKey)
	user, ok := v.(auth.User)
	return user, ok
}

func withUser(ctx context.Context, user auth.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func sessionTokenFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(sessionTokenCtxKey)
	token, ok := v.(string)
	return token, ok
}

func withSessionToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, sessionTokenCtxKey, token)
}

func auditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		userID := "anonymous"
		role := "-"
		if user, ok := userFromContext(r.Context()); ok {
			userID = user.ID
			role = user.Role
		}

		record := map[string]any{
			"ts":          time.Now().UTC().Format(time.RFC3339),
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      rw.status,
			"duration_ms": time.Since(start).Milliseconds(),
			"user_id":     userID,
			"role":        role,
		}

		blob, err := json.Marshal(record)
		if err != nil {
			log.Printf("audit marshal failed: %v", err)
			return
		}
		log.Print(string(blob))
	})
}

func logAuditEvent(ctx context.Context, action, target string, details map[string]any) {
	record := map[string]any{
		"ts":     time.Now().UTC().Format(time.RFC3339),
		"type":   "domain_audit",
		"action": strings.TrimSpace(action),
		"target": strings.TrimSpace(target),
	}
	if user, ok := userFromContext(ctx); ok {
		record["user_id"] = user.ID
		record["role"] = user.Role
		record["actor_email"] = user.Email
	}
	if details != nil {
		record["details"] = details
	}
	blob, err := json.Marshal(record)
	if err != nil {
		log.Printf("domain audit marshal failed: %v", err)
		return
	}
	log.Print(string(blob))
}

func corsMiddleware(allowedOrigins map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if _, ok := allowedOrigins[strings.ToLower(origin)]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			token := ""
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token = strings.TrimSpace(authHeader[len("Bearer "):])
			}
			if token == "" {
				cookie, err := r.Cookie("juriscan_session")
				if err == nil {
					token = strings.TrimSpace(cookie.Value)
				}
			}
			if token == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "auth: missing bearer token")
				return
			}
			user, err := authService.ValidateSession(token)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, err.Error())
				return
			}
			ctx := withUser(r.Context(), user)
			ctx = withSessionToken(ctx, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.TrimSpace(strings.ToLower(role))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := userFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "auth: session required")
				return
			}
			if _, ok := allowed[strings.ToLower(user.Role)]; !ok {
				httpx.WriteError(w, http.StatusForbidden, "rbac: role not allowed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}
