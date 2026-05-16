package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                    string
	HTTPAddr                  string
	DatabaseDriver            string
	DatabaseURL               string
	DBPath                    string
	AllowedOrigins            map[string]struct{}
	LoginTokenEcho            bool
	LeadSLA                   time.Duration
	WhatsAppProvider          string
	WhatsAppWebhookToken      string
	WhatsAppMetaAPIVersion    string
	WhatsAppMetaPhoneNumberID string
	WhatsAppMetaAccessToken   string
	WhatsAppMetaVerifyToken   string
	AILogRetentionDays        int
	PublicationRetentionDays  int

	AdminEmails      map[string]struct{}
	ControllerEmails map[string]struct{}
	LawyerEmails     map[string]struct{}
	CommercialEmails map[string]struct{}
}

func LoadConfigFromEnv() Config {
	appEnv := getenvDefault("APP_ENV", "development")
	httpAddr := getenvDefault("HTTP_ADDR", ":8080")
	databaseDriver := strings.ToLower(getenvDefault("DATABASE_DRIVER", "sqlite"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	dbPath := getenvDefault("DB_PATH", "juriscan.db")
	echo := strings.EqualFold(getenvDefault("LOGIN_TOKEN_ECHO", "true"), "true")
	slaMinutes := parseIntDefault(getenvDefault("LEAD_SLA_MINUTES", "30"), 30)
	aiRetentionDays := parseIntDefault(getenvDefault("AI_LOG_RETENTION_DAYS", "90"), 90)
	publicationRetentionDays := parseIntDefault(getenvDefault("PUBLICATION_RETENTION_DAYS", "365"), 365)

	return Config{
		AppEnv:                    appEnv,
		HTTPAddr:                  httpAddr,
		DatabaseDriver:            databaseDriver,
		DatabaseURL:               databaseURL,
		DBPath:                    dbPath,
		AllowedOrigins:            parseCSVSet(getenvDefault("ALLOWED_ORIGINS", "http://localhost:5174")),
		LoginTokenEcho:            echo,
		LeadSLA:                   time.Duration(slaMinutes) * time.Minute,
		WhatsAppProvider:          normalizeWhatsAppProvider(getenvDefault("WHATSAPP_PROVIDER", "mock")),
		WhatsAppWebhookToken:      strings.TrimSpace(os.Getenv("WHATSAPP_WEBHOOK_TOKEN")),
		WhatsAppMetaAPIVersion:    strings.TrimSpace(getenvDefault("WHATSAPP_META_API_VERSION", "v20.0")),
		WhatsAppMetaPhoneNumberID: strings.TrimSpace(os.Getenv("WHATSAPP_META_PHONE_NUMBER_ID")),
		WhatsAppMetaAccessToken:   strings.TrimSpace(os.Getenv("WHATSAPP_META_ACCESS_TOKEN")),
		WhatsAppMetaVerifyToken:   strings.TrimSpace(os.Getenv("WHATSAPP_META_VERIFY_TOKEN")),
		AILogRetentionDays:        aiRetentionDays,
		PublicationRetentionDays:  publicationRetentionDays,
		AdminEmails:               parseEmailSet(os.Getenv("ADMIN_EMAILS")),
		ControllerEmails:          parseEmailSet(os.Getenv("CONTROLLER_EMAILS")),
		LawyerEmails:              parseEmailSet(os.Getenv("LAWYER_EMAILS")),
		CommercialEmails:          parseEmailSet(os.Getenv("COMMERCIAL_EMAILS")),
	}
}

func normalizeWhatsAppProvider(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "meta_cloud":
		return "meta_cloud"
	default:
		return "mock"
	}
}

func (c Config) ResolveRole(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, ok := c.AdminEmails[email]; ok {
		return "admin"
	}
	if _, ok := c.ControllerEmails[email]; ok {
		return "controller"
	}
	if _, ok := c.LawyerEmails[email]; ok {
		return "lawyer"
	}
	if _, ok := c.CommercialEmails[email]; ok {
		return "commercial"
	}
	return "commercial"
}

func parseEmailSet(raw string) map[string]struct{} {
	return parseCSVSet(raw)
}

func parseCSVSet(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		value := strings.ToLower(strings.TrimSpace(part))
		if value == "" {
			continue
		}
		out[value] = struct{}{}
	}
	return out
}

func getenvDefault(name, fallback string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return fallback
	}
	return v
}

func parseIntDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
