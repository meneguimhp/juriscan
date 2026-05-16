package app

import "time"

type LeadHistoryEvent struct {
	Type      string    `json:"type"`
	At        time.Time `json:"at"`
	FromStage string    `json:"from_stage,omitempty"`
	ToStage   string    `json:"to_stage,omitempty"`
	Note      string    `json:"note,omitempty"`
}

type Lead struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Phone           string             `json:"phone"`
	Origin          string             `json:"origin"`
	Subject         string             `json:"subject"`
	Stage           string             `json:"stage"`
	OwnerEmail      string             `json:"owner_email"`
	NextStep        string             `json:"next_step,omitempty"`
	NextFollowUpAt  *time.Time         `json:"next_follow_up_at,omitempty"`
	AIClassification *LeadAIClassification `json:"ai_classification,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	FirstResponseAt *time.Time         `json:"first_response_at,omitempty"`
	History         []LeadHistoryEvent `json:"history,omitempty"`
}

type LeadAIClassification struct {
	Category        string     `json:"category"`
	Urgency         string     `json:"urgency"`
	Score           int        `json:"score"`
	Justification   string     `json:"justification"`
	SuggestedAction string     `json:"suggested_action"`
	Confidence      float64    `json:"confidence"`
	Model           string     `json:"model"`
	GeneratedAt     time.Time  `json:"generated_at"`
	OverriddenBy    string     `json:"overridden_by,omitempty"`
	OverriddenAt    *time.Time `json:"overridden_at,omitempty"`
	OverrideReason  string     `json:"override_reason,omitempty"`
}

type LeadQueueItem struct {
	Lead
	MinutesWithoutResponse int64      `json:"minutes_without_response"`
	SLAStatus              string     `json:"sla_status"`
	ResponseDueAt          *time.Time `json:"response_due_at,omitempty"`
}

type WhatsAppConversation struct {
	ID            string    `json:"id"`
	Phone         string    `json:"phone"`
	ContactName   string    `json:"contact_name,omitempty"`
	LastMessage   string    `json:"last_message"`
	Status        string    `json:"status"`
	LeadID        string    `json:"lead_id,omitempty"`
	SuggestedLeadID   string    `json:"suggested_lead_id,omitempty"`
	SuggestedLeadName string    `json:"suggested_lead_name,omitempty"`
	MessageCount  int       `json:"message_count"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WhatsAppInboundMessage struct {
	Phone       string    `json:"phone"`
	Message     string    `json:"message"`
	ContactName string    `json:"contact_name"`
	ExternalID  string    `json:"external_id"`
	ReceivedAt  time.Time `json:"received_at"`
}

type MessageTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Channel   string    `json:"channel"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FollowUp struct {
	ID         string     `json:"id"`
	LeadID     string     `json:"lead_id"`
	TemplateID string     `json:"template_id,omitempty"`
	Message    string     `json:"message"`
	DueAt      time.Time  `json:"due_at"`
	Status     string     `json:"status"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DoneAt     *time.Time `json:"done_at,omitempty"`
}

type Publication struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"`
	InputType  string                 `json:"input_type"`
	FileName   string                 `json:"file_name,omitempty"`
	RawText    string                 `json:"raw_text"`
	CreatedBy  string                 `json:"created_by"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Analysis   *PublicationAnalysis   `json:"analysis,omitempty"`
	Validation *PublicationValidation `json:"validation,omitempty"`
	TaskID     string                 `json:"task_id,omitempty"`
}

type PublicationAnalysis struct {
	ActType               string    `json:"act_type"`
	BaseDate              time.Time `json:"base_date"`
	SuggestedDeadlineDays int       `json:"suggested_deadline_days"`
	SuggestedDeadlineAt   time.Time `json:"suggested_deadline_at"`
	Risk                  string    `json:"risk"`
	SuggestedOwnerEmail   string    `json:"suggested_owner_email"`
	Confidence            float64   `json:"confidence"`
	Model                 string    `json:"model"`
	Prompt                string    `json:"prompt"`
	Response              string    `json:"response"`
	GeneratedAt           time.Time `json:"generated_at"`
}

type PublicationValidation struct {
	ValidatedBy     string    `json:"validated_by"`
	ValidatedAt     time.Time `json:"validated_at"`
	FinalDeadlineAt time.Time `json:"final_deadline_at"`
	Notes           string    `json:"notes,omitempty"`
}

type DeadlineTask struct {
	ID            string    `json:"id"`
	PublicationID string    `json:"publication_id"`
	Title         string    `json:"title"`
	OwnerEmail    string    `json:"owner_email"`
	DueAt         time.Time `json:"due_at"`
	Status        string    `json:"status"`
	Risk          string    `json:"risk"`
	Checklist     []string  `json:"checklist,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeadlineAlert struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type AILogRecord struct {
	ID             string    `json:"id"`
	Feature        string    `json:"feature"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     string    `json:"resource_id"`
	Prompt         string    `json:"prompt"`
	Response       string    `json:"response"`
	Model          string    `json:"model"`
	Confidence     float64   `json:"confidence"`
	CreatedAt      time.Time `json:"created_at"`
	RetentionUntil time.Time `json:"retention_until"`
}
