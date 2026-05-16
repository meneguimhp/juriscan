package app

import (
	"net/url"
	"testing"
)

func TestParseMetaWebhookMessages(t *testing.T) {
	payload := []byte(`{
		"entry":[
			{
				"changes":[
					{
						"value":{
							"contacts":[
								{"wa_id":"5511999990000","profile":{"name":"Cliente Juriscan"}}
							],
							"messages":[
								{
									"id":"wamid.HBgMNTUxMTk5OTk5MDAwMA==",
									"timestamp":"1776180487",
									"from":"5511999990000",
									"type":"text",
									"text":{"body":"Preciso de orientacao"}
								}
							]
						}
					}
				]
			}
		]
	}`)

	items, err := parseMetaWebhookMessages(payload)
	if err != nil {
		t.Fatalf("parseMetaWebhookMessages returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 message, got %d", len(items))
	}
	if items[0].Phone != "5511999990000" {
		t.Fatalf("unexpected phone: %s", items[0].Phone)
	}
	if items[0].ContactName != "Cliente Juriscan" {
		t.Fatalf("unexpected contact name: %s", items[0].ContactName)
	}
	if items[0].Message != "Preciso de orientacao" {
		t.Fatalf("unexpected message: %s", items[0].Message)
	}
}

func TestVerifyMetaWebhook(t *testing.T) {
	cfg := Config{WhatsAppMetaVerifyToken: "token-seguro"}
	query := url.Values{
		"hub.mode":         []string{"subscribe"},
		"hub.verify_token": []string{"token-seguro"},
		"hub.challenge":    []string{"challenge-ok"},
	}

	challenge, err := verifyMetaWebhook(cfg, query)
	if err != nil {
		t.Fatalf("verifyMetaWebhook returned error: %v", err)
	}
	if challenge != "challenge-ok" {
		t.Fatalf("unexpected challenge: %s", challenge)
	}
}

func TestSendMetaCloudTextRequiresConfig(t *testing.T) {
	cfg := Config{
		WhatsAppProvider: "meta_cloud",
	}
	if _, err := sendMetaCloudText(cfg, "5511999990000", "Teste"); err == nil {
		t.Fatalf("expected error for missing configuration")
	}
}
