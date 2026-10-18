package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestVerifyCaptchaIncludesWidgetRecordID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if got := payload["widget_record_id"]; got != "widget-123" {
			t.Errorf("widget_record_id = %v, want widget-123", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer server.Close()

	client := &InfraiClient{
		key: "test-key", widgetRecordID: "widget-123", http: server.Client(),
		sleep: func(_ time.Duration) {}, baseURL: server.URL,
	}
	valid, err := client.VerifyCaptcha(context.Background(), "token-123", "192.0.2.1")
	if err != nil || !valid {
		t.Fatalf("VerifyCaptcha() = %v, %v", valid, err)
	}
}

func TestRouteCreatorContactRecordsConsentAndEscapesMessage(t *testing.T) {
	tests := []struct {
		name    string
		form    ContactSubmission
		wantErr bool
	}{
		{"accepted consent", ContactSubmission{Email: "artist@example.com", Name: "Ari", Message: "License <details>", Consent: true}, false},
		{"missing consent", ContactSubmission{Email: "artist@example.com", Message: "Hello", Consent: false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mail, err := routeCreatorContact(tt.form, "team@example.com")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (!strings.Contains(mail.HTML, "Consent:</b> recorded") || !strings.Contains(mail.HTML, "&lt;details&gt;")) {
				t.Fatalf("mail body did not preserve the contact decision: %s", mail.HTML)
			}
		})
	}
}
