package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func parseMetaWebhookMessages(body []byte) ([]WhatsAppInboundMessage, error) {
	var payload struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						From      string `json:"from"`
						Type      string `json:"type"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
					} `json:"messages"`
					Contacts []struct {
						WAID    string `json:"wa_id"`
						Profile struct {
							Name string `json:"name"`
						} `json:"profile"`
					} `json:"contacts"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	contactByPhone := make(map[string]string, 16)
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, contact := range change.Value.Contacts {
				phone := normalizePhone(contact.WAID)
				name := strings.TrimSpace(contact.Profile.Name)
				if phone != "" && name != "" {
					contactByPhone[phone] = name
				}
			}
		}
	}

	items := make([]WhatsAppInboundMessage, 0, 8)
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				phone := normalizePhone(msg.From)
				if phone == "" {
					continue
				}
				if strings.TrimSpace(strings.ToLower(msg.Type)) != "text" {
					continue
				}
				text := strings.TrimSpace(msg.Text.Body)
				if text == "" {
					continue
				}
				receivedAt := time.Now().UTC()
				if ts := strings.TrimSpace(msg.Timestamp); ts != "" {
					if sec, err := parseUnixSeconds(ts); err == nil {
						receivedAt = sec
					}
				}
				items = append(items, WhatsAppInboundMessage{
					Phone:       phone,
					Message:     text,
					ContactName: contactByPhone[phone],
					ExternalID:  strings.TrimSpace(msg.ID),
					ReceivedAt:  receivedAt,
				})
			}
		}
	}

	if len(items) == 0 {
		return nil, errors.New("whatsapp: no valid messages in meta payload")
	}
	return items, nil
}

func verifyMetaWebhook(cfg Config, query map[string][]string) (string, error) {
	mode := strings.TrimSpace(queryValue(query, "hub.mode"))
	token := strings.TrimSpace(queryValue(query, "hub.verify_token"))
	challenge := strings.TrimSpace(queryValue(query, "hub.challenge"))
	if mode != "subscribe" || challenge == "" {
		return "", errors.New("whatsapp: invalid verification query")
	}
	if strings.TrimSpace(cfg.WhatsAppMetaVerifyToken) == "" || token != strings.TrimSpace(cfg.WhatsAppMetaVerifyToken) {
		return "", errors.New("whatsapp: invalid verify token")
	}
	return challenge, nil
}

func sendMetaCloudText(cfg Config, to, message string) (map[string]any, error) {
	to = normalizePhone(to)
	message = strings.TrimSpace(message)
	if to == "" || message == "" {
		return nil, errors.New("whatsapp: invalid outbound payload")
	}
	if cfg.WhatsAppMetaPhoneNumberID == "" || cfg.WhatsAppMetaAccessToken == "" {
		return nil, errors.New("whatsapp: missing meta configuration")
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", cfg.WhatsAppMetaAPIVersion, cfg.WhatsAppMetaPhoneNumberID)
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]any{
			"preview_url": false,
			"body":        message,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.WhatsAppMetaAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("whatsapp: meta send failed: %s - %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	out := map[string]any{}
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &out)
	}
	out["provider"] = "meta_cloud"
	out["to"] = to
	return out, nil
}

func parseUnixSeconds(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("empty timestamp")
	}
	var sec int64
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return time.Time{}, errors.New("invalid timestamp")
		}
		sec = sec*10 + int64(ch-'0')
	}
	return time.Unix(sec, 0).UTC(), nil
}

func queryValue(query map[string][]string, key string) string {
	values := query[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
