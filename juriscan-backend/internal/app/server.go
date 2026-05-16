package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"juriscan-backend/internal/httpx"
	"juriscan-backend/internal/identity/auth"
)

const sessionCookieName = "juriscan_session"

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	db, err := openApplicationDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("user db: %w", err)
	}
	userStore, err := NewUserStore(db)
	if err != nil {
		return nil, fmt.Errorf("user store: %w", err)
	}

	for email := range cfg.ControllerEmails {
		_ = userStore.EnsureBootstrapUser(deriveNameFromEmail(email), email, "controller")
	}
	for email := range cfg.LawyerEmails {
		_ = userStore.EnsureBootstrapUser(deriveNameFromEmail(email), email, "lawyer")
	}
	for email := range cfg.CommercialEmails {
		_ = userStore.EnsureBootstrapUser(deriveNameFromEmail(email), email, "commercial")
	}
	for email := range cfg.AdminEmails {
		_ = userStore.EnsureBootstrapUser(deriveNameFromEmail(email), email, "admin")
	}

	if len(userStore.List()) == 0 {
		_ = userStore.EnsureBootstrapUser("Admin", "admin@juriscan.local", "admin")
	}
	authService := auth.NewService(10*time.Minute, 24*time.Hour, userStore.ResolveAuthUser)
	leadStore := NewLeadStore()
	conversationStore := NewConversationStore()
	templateFollowUpStore := NewTemplateFollowUpStore()
	publicationStore := NewPublicationStore()
	deadlineStore := NewDeadlineStore()
	aiLogStore := NewAILogStore()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	mux.HandleFunc("POST /v1/identity/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "auth: invalid payload")
			return
		}

		token, expiresIn, err := authService.RequestOTP(payload.Email)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		resp := map[string]any{
			"message":            "otp_sent",
			"expires_in_seconds": expiresIn,
		}
		if cfg.LoginTokenEcho {
			resp["token"] = token
		}
		httpx.WriteJSON(w, http.StatusAccepted, resp)
	})

	mux.HandleFunc("POST /v1/identity/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Email string `json:"email"`
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "auth: invalid payload")
			return
		}

		accessToken, user, err := authService.VerifyOTP(payload.Email, payload.Token)
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, auth.ErrInvalidEmail) {
				status = http.StatusBadRequest
			}
			httpx.WriteError(w, status, err.Error())
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    accessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.AppEnv != "development",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int((24 * time.Hour).Seconds()),
		})
		userStore.MarkLogin(user.ID, time.Now().UTC())

		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"access_token": accessToken,
			"user":         user,
		})
	})

	protected := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _ := userFromContext(r.Context())
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"user": user,
			})
		}),
		authMiddleware(authService),
	)
	mux.Handle("GET /v1/identity/me", protected)

	logoutRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			if token, ok := sessionTokenFromContext(r.Context()); ok {
				authService.RevokeSession(token)
			}
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.AppEnv != "development",
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			})
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": "logged_out"})
		}),
		authMiddleware(authService),
	)
	mux.Handle("/v1/identity/logout", logoutRoute)

	leadsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, _ := userFromContext(r.Context())
		switch r.Method {
		case http.MethodGet:
			rawFilter := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("sla")))
			onlyOverdue := rawFilter == LeadSLAStatusOverdue || rawFilter == "overdue"
			items := BuildLeadQueue(leadStore.List(), time.Now().UTC(), cfg.LeadSLA, onlyOverdue)
			for i := range items {
				items[i].Lead = maskLeadByRole(items[i].Lead, currentUser.Role)
			}
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"items":       items,
				"sla_minutes": int(cfg.LeadSLA.Minutes()),
			})
		case http.MethodPost:
			if !hasRole(currentUser.Role, "admin", "commercial") {
				httpx.WriteError(w, http.StatusForbidden, "rbac: role not allowed")
				return
			}
			var payload Lead
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "lead: invalid payload")
				return
			}
			payload.OwnerEmail = strings.TrimSpace(payload.OwnerEmail)
			item, err := leadStore.Create(payload)
			if err != nil {
				status := http.StatusBadRequest
				if !errors.Is(err, ErrInvalidLead) {
					status = http.StatusInternalServerError
				}
				httpx.WriteError(w, status, err.Error())
				return
			}
			logAuditEvent(r.Context(), "lead.created", item.ID, map[string]any{
				"origin": item.Origin,
				"stage":  item.Stage,
			})
			httpx.WriteJSON(w, http.StatusCreated, item)
		default:
			httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
		}
	})

	crmRoute := chain(
		leadsHandler,
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/leads", crmRoute)

	leadEditHandler := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "lead: id required")
				return
			}

			var payload struct {
				Name       *string `json:"name"`
				Phone      *string `json:"phone"`
				Origin     *string `json:"origin"`
				Subject    *string `json:"subject"`
				OwnerEmail *string `json:"owner_email"`
				Stage      *string `json:"stage"`
				NextStep   *string `json:"next_step"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "lead: invalid payload")
				return
			}

			updated, err := leadStore.Update(id, LeadUpdate{
				Name:       payload.Name,
				Phone:      payload.Phone,
				Origin:     payload.Origin,
				Subject:    payload.Subject,
				OwnerEmail: payload.OwnerEmail,
				Stage:      payload.Stage,
				NextStep:   payload.NextStep,
			})
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidLead):
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				case errors.Is(err, ErrLeadNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "lead: unexpected error")
				}
				return
			}

			logAuditEvent(r.Context(), "lead.updated", updated.ID, map[string]any{
				"stage":     updated.Stage,
				"next_step": updated.NextStep,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/leads/{id}", leadEditHandler)

	leadStageHandler := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "lead: id required")
				return
			}
			var payload struct {
				Stage string `json:"stage"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "lead: invalid payload")
				return
			}

			updated, err := leadStore.UpdateStage(id, payload.Stage)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidLead):
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				case errors.Is(err, ErrLeadNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "lead: unexpected error")
				}
				return
			}

			logAuditEvent(r.Context(), "lead.stage_changed", updated.ID, map[string]any{
				"stage": updated.Stage,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/leads/{id}/stage", leadStageHandler)

	leadTriageRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			leadID := strings.TrimSpace(r.PathValue("id"))
			if leadID == "" {
				httpx.WriteError(w, http.StatusBadRequest, "lead: id required")
				return
			}

			lead, err := leadStore.GetByID(leadID)
			if err != nil {
				if errors.Is(err, ErrLeadNotFound) {
					httpx.WriteError(w, http.StatusNotFound, err.Error())
					return
				}
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			classification := classifyLeadHeuristic(lead)
			prompt, response := buildLeadTriagePromptAndResponse(lead, classification)
			updated, err := leadStore.ApplyAIClassification(leadID, classification)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			aiLog := aiLogStore.Add(AILogRecord{
				Feature:        "lead_triage",
				ResourceType:   "lead",
				ResourceID:     leadID,
				Prompt:         prompt,
				Response:       response,
				Model:          classification.Model,
				Confidence:     classification.Confidence,
				RetentionUntil: time.Now().UTC().AddDate(0, 0, cfg.AILogRetentionDays),
			})
			logAuditEvent(r.Context(), "lead.ai_triaged", leadID, map[string]any{
				"category":   classification.Category,
				"urgency":    classification.Urgency,
				"score":      classification.Score,
				"next_step":  classification.SuggestedAction,
				"ai_log_id":  aiLog.ID,
				"confidence": classification.Confidence,
			})

			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"lead":   updated,
				"ai_log": aiLog,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/leads/{id}/triage", leadTriageRoute)

	leadTriageOverrideRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			leadID := strings.TrimSpace(r.PathValue("id"))
			if leadID == "" {
				httpx.WriteError(w, http.StatusBadRequest, "lead: id required")
				return
			}
			currentUser, _ := userFromContext(r.Context())
			var payload struct {
				Reason          string `json:"reason"`
				Category        string `json:"category"`
				Urgency         string `json:"urgency"`
				Score           int    `json:"score"`
				Justification   string `json:"justification"`
				SuggestedAction string `json:"suggested_action"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "lead: invalid payload")
				return
			}

			override := LeadAIClassification{
				Category:        payload.Category,
				Urgency:         payload.Urgency,
				Score:           payload.Score,
				Justification:   payload.Justification,
				SuggestedAction: payload.SuggestedAction,
				Confidence:      1,
				Model:           "human_override",
			}
			updated, err := leadStore.OverrideAIClassification(leadID, payload.Reason, currentUser.Email, override)
			if err != nil {
				switch {
				case errors.Is(err, ErrLeadNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				}
				return
			}
			logAuditEvent(r.Context(), "lead.ai_override", leadID, map[string]any{
				"reason":   strings.TrimSpace(payload.Reason),
				"by":       currentUser.Email,
				"category": updated.AIClassification.Category,
				"urgency":  updated.AIClassification.Urgency,
				"score":    updated.AIClassification.Score,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/leads/{id}/triage/override", leadTriageOverrideRoute)

	templatesRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				httpx.WriteJSON(w, http.StatusOK, map[string]any{
					"items": templateFollowUpStore.ListTemplates(),
				})
			case http.MethodPost:
				currentUser, _ := userFromContext(r.Context())
				if !hasRole(currentUser.Role, "admin", "commercial") {
					httpx.WriteError(w, http.StatusForbidden, "rbac: role not allowed")
					return
				}
				var payload struct {
					Name    string `json:"name"`
					Channel string `json:"channel"`
					Body    string `json:"body"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "template: invalid payload")
					return
				}
				created, err := templateFollowUpStore.CreateTemplate(MessageTemplate{
					Name:    payload.Name,
					Channel: payload.Channel,
					Body:    payload.Body,
				})
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
					return
				}
				logAuditEvent(r.Context(), "template.created", created.ID, map[string]any{
					"name":    created.Name,
					"channel": created.Channel,
				})
				httpx.WriteJSON(w, http.StatusCreated, created)
			default:
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
			}
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/templates", templatesRoute)

	followUpsRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				leadID := strings.TrimSpace(r.URL.Query().Get("lead_id"))
				onlyPending := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("pending")), "true")
				items := templateFollowUpStore.ListFollowUps(leadID, onlyPending)
				httpx.WriteJSON(w, http.StatusOK, map[string]any{
					"items": items,
				})
			case http.MethodPost:
				currentUser, _ := userFromContext(r.Context())
				var payload struct {
					LeadID     string `json:"lead_id"`
					TemplateID string `json:"template_id"`
					Message    string `json:"message"`
					DueAt      string `json:"due_at"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "followup: invalid payload")
					return
				}
				lead, err := leadStore.GetByID(payload.LeadID)
				if err != nil {
					httpx.WriteError(w, http.StatusNotFound, "followup: lead not found")
					return
				}

				message := strings.TrimSpace(payload.Message)
				if strings.TrimSpace(payload.TemplateID) != "" {
					tpl, err := templateFollowUpStore.ResolveTemplate(payload.TemplateID)
					if err != nil {
						httpx.WriteError(w, http.StatusBadRequest, err.Error())
						return
					}
					if message == "" {
						message = tpl.Body
					}
				}
				dueAt, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.DueAt))
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "followup: invalid due_at (use RFC3339)")
					return
				}

				created, err := templateFollowUpStore.CreateFollowUp(FollowUp{
					LeadID:     payload.LeadID,
					TemplateID: payload.TemplateID,
					Message:    message,
					DueAt:      dueAt,
					CreatedBy:  currentUser.Email,
				})
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
					return
				}
				_, _ = leadStore.SetNextFollowUp(lead.ID, dueAt)
				logAuditEvent(r.Context(), "followup.created", created.ID, map[string]any{
					"lead_id": payload.LeadID,
					"due_at":  created.DueAt.Format(time.RFC3339),
				})
				httpx.WriteJSON(w, http.StatusCreated, created)
			default:
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
			}
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/followups", followUpsRoute)

	followUpUpdateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "followup: id required")
				return
			}
			var payload struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "followup: invalid payload")
				return
			}
			updated, err := templateFollowUpStore.UpdateFollowUpStatus(id, payload.Status)
			if err != nil {
				switch {
				case errors.Is(err, ErrFollowUpNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				}
				return
			}
			logAuditEvent(r.Context(), "followup.updated", updated.ID, map[string]any{
				"status": updated.Status,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial", "controller", "lawyer"),
	)
	mux.Handle("/v1/crm/followups/{id}", followUpUpdateRoute)

	whatsAppWebhook := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if cfg.WhatsAppProvider != "meta_cloud" {
				httpx.WriteError(w, http.StatusNotFound, "whatsapp: webhook verification only for meta_cloud")
				return
			}
			challenge, err := verifyMetaWebhook(cfg, r.URL.Query())
			if err != nil {
				httpx.WriteError(w, http.StatusForbidden, err.Error())
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(challenge))
			return

		case http.MethodPost:
			body, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024))
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "whatsapp: invalid payload")
				return
			}

			messages := make([]WhatsAppInboundMessage, 0, 4)
			switch cfg.WhatsAppProvider {
			case "meta_cloud":
				parsed, parseErr := parseMetaWebhookMessages(body)
				if parseErr != nil {
					httpx.WriteError(w, http.StatusBadRequest, parseErr.Error())
					return
				}
				messages = parsed
			default:
				if cfg.WhatsAppWebhookToken != "" {
					if strings.TrimSpace(r.Header.Get("X-Webhook-Token")) != cfg.WhatsAppWebhookToken {
						httpx.WriteError(w, http.StatusUnauthorized, "whatsapp: invalid webhook token")
						return
					}
				}
				var payload WhatsAppInboundMessage
				if err := json.Unmarshal(body, &payload); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "whatsapp: invalid payload")
					return
				}
				messages = append(messages, payload)
			}

			accepted := make([]WhatsAppConversation, 0, len(messages))
			for _, incoming := range messages {
				created, ingestErr := conversationStore.Ingest(incoming)
				if ingestErr != nil {
					status := http.StatusBadRequest
					if !errors.Is(ingestErr, ErrInvalidConversation) {
						status = http.StatusInternalServerError
					}
					httpx.WriteError(w, status, ingestErr.Error())
					return
				}
				logAuditEvent(r.Context(), "whatsapp.message_ingested", created.ID, map[string]any{
					"phone":         created.Phone,
					"status":        created.Status,
					"message_count": created.MessageCount,
				})
				accepted = append(accepted, created)
			}

			httpx.WriteJSON(w, http.StatusAccepted, map[string]any{
				"provider": cfg.WhatsAppProvider,
				"accepted": len(accepted),
				"items":    accepted,
			})
		default:
			httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
		}
	})
	mux.Handle("/v1/whatsapp/webhook", whatsAppWebhook)

	whatsAppSimulateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			var payload WhatsAppInboundMessage
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "whatsapp: invalid payload")
				return
			}
			created, err := conversationStore.Ingest(payload)
			if err != nil {
				status := http.StatusBadRequest
				if !errors.Is(err, ErrInvalidConversation) {
					status = http.StatusInternalServerError
				}
				httpx.WriteError(w, status, err.Error())
				return
			}
			logAuditEvent(r.Context(), "whatsapp.simulated_message_ingested", created.ID, map[string]any{
				"phone":         created.Phone,
				"status":        created.Status,
				"message_count": created.MessageCount,
			})
			httpx.WriteJSON(w, http.StatusAccepted, created)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial"),
	)
	mux.Handle("/v1/whatsapp/simulate", whatsAppSimulateRoute)

	whatsAppSendRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			var payload struct {
				To      string `json:"to"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "whatsapp: invalid payload")
				return
			}
			var result map[string]any
			var err error
			switch cfg.WhatsAppProvider {
			case "meta_cloud":
				result, err = sendMetaCloudText(cfg, payload.To, payload.Message)
			default:
				result = map[string]any{
					"provider": "mock",
					"to":       normalizePhone(payload.To),
					"message":  strings.TrimSpace(payload.Message),
					"status":   "simulated_sent",
				}
			}
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			logAuditEvent(r.Context(), "whatsapp.outbound_sent", "outbound", result)
			httpx.WriteJSON(w, http.StatusAccepted, result)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial"),
	)
	mux.Handle("/v1/whatsapp/messages/send", whatsAppSendRoute)

	whatsAppConversationsRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			currentUser, _ := userFromContext(r.Context())
			items := conversationStore.List()
			for i := range items {
				if matches := leadStore.FindByPhone(items[i].Phone); len(matches) > 0 {
					items[i].SuggestedLeadID = matches[0].ID
					items[i].SuggestedLeadName = matches[0].Name
				}
				items[i] = maskConversationByRole(items[i], currentUser.Role)
			}
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"items": items,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial"),
	)
	mux.Handle("/v1/whatsapp/conversations", whatsAppConversationsRoute)

	whatsAppConversationUpdateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "whatsapp: id required")
				return
			}
			var payload struct {
				Status string `json:"status"`
				LeadID string `json:"lead_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "whatsapp: invalid payload")
				return
			}
			if strings.TrimSpace(payload.LeadID) != "" {
				if _, err := leadStore.GetByID(payload.LeadID); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "whatsapp: lead not found")
					return
				}
			}
			updated, err := conversationStore.UpdateClassification(id, payload.Status, payload.LeadID)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidConversation):
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				case errors.Is(err, ErrConversationNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "whatsapp: unexpected error")
				}
				return
			}
			logAuditEvent(r.Context(), "whatsapp.classification_updated", updated.ID, map[string]any{
				"status":  updated.Status,
				"lead_id": updated.LeadID,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "commercial"),
	)
	mux.Handle("/v1/whatsapp/conversations/{id}", whatsAppConversationUpdateRoute)

	publicationsRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				currentUser, _ := userFromContext(r.Context())
				items := publicationStore.List()
				for i := range items {
					items[i] = maskPublicationByRole(items[i], currentUser.Role)
				}
				httpx.WriteJSON(w, http.StatusOK, map[string]any{
					"items": items,
				})
			case http.MethodPost:
				currentUser, _ := userFromContext(r.Context())
				var payload struct {
					Source    string `json:"source"`
					InputType string `json:"input_type"`
					FileName  string `json:"file_name"`
					RawText   string `json:"raw_text"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "publication: invalid payload")
					return
				}
				created, err := publicationStore.Create(Publication{
					Source:    payload.Source,
					InputType: payload.InputType,
					FileName:  payload.FileName,
					RawText:   payload.RawText,
					CreatedBy: currentUser.Email,
				})
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
					return
				}
				logAuditEvent(r.Context(), "publication.created", created.ID, map[string]any{
					"source": created.Source,
					"type":   created.InputType,
				})
				httpx.WriteJSON(w, http.StatusCreated, created)
			default:
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
			}
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/publications", publicationsRoute)

	publicationAnalyzeRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "publication: id required")
				return
			}
			publication, err := publicationStore.GetByID(id)
			if err != nil {
				httpx.WriteError(w, http.StatusNotFound, err.Error())
				return
			}

			analysis := analyzePublicationHeuristic(publication)
			updated, err := publicationStore.SetAnalysis(id, analysis)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			aiLog := aiLogStore.Add(AILogRecord{
				Feature:        "publication_analysis",
				ResourceType:   "publication",
				ResourceID:     updated.ID,
				Prompt:         analysis.Prompt,
				Response:       analysis.Response,
				Model:          analysis.Model,
				Confidence:     analysis.Confidence,
				RetentionUntil: time.Now().UTC().AddDate(0, 0, cfg.AILogRetentionDays),
			})
			logAuditEvent(r.Context(), "publication.analyzed", updated.ID, map[string]any{
				"risk":      analysis.Risk,
				"deadline":  analysis.SuggestedDeadlineAt.Format(time.RFC3339),
				"ai_log_id": aiLog.ID,
			})
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"publication": updated,
				"ai_log":      aiLog,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/publications/{id}/analyze", publicationAnalyzeRoute)

	publicationValidateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "publication: id required")
				return
			}
			currentUser, _ := userFromContext(r.Context())
			var payload struct {
				FinalDeadlineAt string `json:"final_deadline_at"`
				Notes           string `json:"notes"`
				OwnerEmail      string `json:"owner_email"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "publication: invalid payload")
				return
			}
			finalDeadline, err := time.Parse(time.RFC3339, strings.TrimSpace(payload.FinalDeadlineAt))
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "publication: invalid final_deadline_at (use RFC3339)")
				return
			}
			validated, err := publicationStore.Validate(id, PublicationValidation{
				ValidatedBy:     currentUser.Email,
				ValidatedAt:     time.Now().UTC(),
				FinalDeadlineAt: finalDeadline,
				Notes:           payload.Notes,
			})
			if err != nil {
				switch {
				case errors.Is(err, ErrPublicationNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				}
				return
			}

			ownerEmail := strings.TrimSpace(strings.ToLower(payload.OwnerEmail))
			if ownerEmail == "" && validated.Analysis != nil {
				ownerEmail = validated.Analysis.SuggestedOwnerEmail
			}
			if ownerEmail == "" {
				ownerEmail = currentUser.Email
			}
			risk := "medio"
			if validated.Analysis != nil {
				risk = validated.Analysis.Risk
			}
			title := "Prazo processual"
			if validated.Analysis != nil && validated.Analysis.ActType != "" {
				title = "Prazo de " + validated.Analysis.ActType
			}
			task, err := deadlineStore.CreateTask(DeadlineTask{
				PublicationID: validated.ID,
				Title:         title,
				OwnerEmail:    ownerEmail,
				DueAt:         validated.Validation.FinalDeadlineAt,
				Risk:          risk,
				Checklist: []string{
					"Conferir ato processual",
					"Validar estrategia com advogado responsavel",
					"Registrar protocolo/entrega",
				},
			})
			if err == nil {
				validated, _ = publicationStore.AttachTask(validated.ID, task.ID)
			}
			logAuditEvent(r.Context(), "publication.validated", validated.ID, map[string]any{
				"final_deadline": validated.Validation.FinalDeadlineAt.Format(time.RFC3339),
				"task_id":        validated.TaskID,
				"owner_email":    ownerEmail,
			})
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"publication": validated,
				"task":        task,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/publications/{id}/validate", publicationValidateRoute)

	deadlineTasksRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				owner := strings.TrimSpace(r.URL.Query().Get("owner_email"))
				status := strings.TrimSpace(r.URL.Query().Get("status"))
				items := deadlineStore.ListTasks(owner, status)
				httpx.WriteJSON(w, http.StatusOK, map[string]any{
					"items": items,
				})
			default:
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
			}
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/deadlines/tasks", deadlineTasksRoute)

	deadlineTaskUpdateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "deadline: id required")
				return
			}
			var payload struct {
				Status     string `json:"status"`
				OwnerEmail string `json:"owner_email"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "deadline: invalid payload")
				return
			}
			updated, err := deadlineStore.UpdateTask(id, payload.Status, payload.OwnerEmail)
			if err != nil {
				switch {
				case errors.Is(err, ErrDeadlineTaskNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				default:
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				}
				return
			}
			logAuditEvent(r.Context(), "deadline.task_updated", updated.ID, map[string]any{
				"status":      updated.Status,
				"owner_email": updated.OwnerEmail,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller"),
	)
	mux.Handle("/v1/deadlines/tasks/{id}", deadlineTaskUpdateRoute)

	deadlineAlertsRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			alerts := deadlineStore.BuildAlerts(time.Now().UTC())
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"items": alerts,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller", "lawyer"),
	)
	mux.Handle("/v1/deadlines/alerts", deadlineAlertsRoute)

	complianceAILogsRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			feature := strings.TrimSpace(r.URL.Query().Get("feature"))
			items := aiLogStore.List(feature)
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"items": items,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin", "controller"),
	)
	mux.Handle("/v1/compliance/ai-logs", complianceAILogsRoute)

	complianceRetentionRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			now := time.Now().UTC()
			removedAILogs := aiLogStore.PurgeExpired(now)
			removedPublications := publicationStore.PurgeOlderThan(now.AddDate(0, 0, -cfg.PublicationRetentionDays))

			logAuditEvent(r.Context(), "compliance.retention_run", "retention", map[string]any{
				"removed_ai_logs":      removedAILogs,
				"removed_publications": removedPublications,
			})
			httpx.WriteJSON(w, http.StatusOK, map[string]any{
				"removed_ai_logs":      removedAILogs,
				"removed_publications": removedPublications,
			})
		}),
		authMiddleware(authService),
		requireRoles("admin"),
	)
	mux.Handle("/v1/compliance/retention/run", complianceRetentionRoute)

	adminUsersRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				httpx.WriteJSON(w, http.StatusOK, map[string]any{
					"items": userStore.List(),
				})
			case http.MethodPost:
				var payload struct {
					Name   string `json:"name"`
					Email  string `json:"email"`
					Role   string `json:"role"`
					Status string `json:"status"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					httpx.WriteError(w, http.StatusBadRequest, "user: invalid payload")
					return
				}

				created, err := userStore.Create(UserRecord{
					Name:   payload.Name,
					Email:  payload.Email,
					Role:   payload.Role,
					Status: payload.Status,
				})
				if err != nil {
					switch {
					case errors.Is(err, ErrUserAlreadyExists):
						httpx.WriteError(w, http.StatusConflict, err.Error())
					case errors.Is(err, ErrInvalidUserPayload):
						httpx.WriteError(w, http.StatusBadRequest, err.Error())
					default:
						httpx.WriteError(w, http.StatusInternalServerError, "user: unexpected error")
					}
					return
				}

				logAuditEvent(r.Context(), "user.created", created.ID, map[string]any{
					"email":  created.Email,
					"role":   created.Role,
					"status": created.Status,
				})
				httpx.WriteJSON(w, http.StatusCreated, created)
			default:
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
			}
		}),
		authMiddleware(authService),
		requireRoles("admin"),
	)
	mux.Handle("/v1/admin/users", adminUsersRoute)

	adminUserUpdateRoute := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				httpx.WriteError(w, http.StatusMethodNotAllowed, fmt.Sprintf("%s not allowed", r.Method))
				return
			}
			id := strings.TrimSpace(r.PathValue("id"))
			if id == "" {
				httpx.WriteError(w, http.StatusBadRequest, "user: id required")
				return
			}
			var payload struct {
				Name   *string `json:"name"`
				Role   *string `json:"role"`
				Status *string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "user: invalid payload")
				return
			}

			updated, err := userStore.Update(id, UserUpdate{
				Name:   payload.Name,
				Role:   payload.Role,
				Status: payload.Status,
			})
			if err != nil {
				switch {
				case errors.Is(err, ErrUserNotFound):
					httpx.WriteError(w, http.StatusNotFound, err.Error())
				case errors.Is(err, ErrInvalidUserPayload):
					httpx.WriteError(w, http.StatusBadRequest, err.Error())
				default:
					httpx.WriteError(w, http.StatusInternalServerError, "user: unexpected error")
				}
				return
			}
			logAuditEvent(r.Context(), "user.updated", updated.ID, map[string]any{
				"email":  updated.Email,
				"role":   updated.Role,
				"status": updated.Status,
			})
			httpx.WriteJSON(w, http.StatusOK, updated)
		}),
		authMiddleware(authService),
		requireRoles("admin"),
	)
	mux.Handle("/v1/admin/users/{id}", adminUserUpdateRoute)

	finalHandler := auditMiddleware(corsMiddleware(cfg.AllowedOrigins, mux))
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           finalHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &Server{httpServer: server}, nil
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func hasRole(current string, allowed ...string) bool {
	current = strings.TrimSpace(strings.ToLower(current))
	if current == "" {
		return false
	}
	for _, role := range allowed {
		if current == strings.TrimSpace(strings.ToLower(role)) {
			return true
		}
	}
	return false
}
