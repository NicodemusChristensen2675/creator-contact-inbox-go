package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	inbox := os.Getenv("TEAM_INBOX_EMAIL")
	if inbox == "" {
		log.Fatal("TEAM_INBOX_EMAIL is required")
	}
	http.HandleFunc("/contact", contactHandler(client, inbox))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func contactHandler(client *InfraiClient, inbox string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var form ContactSubmission
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		mail, err := routeCreatorContact(form, inbox)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		valid, err := client.VerifyCaptcha(r.Context(), form.CaptchaToken, ip)
		if err != nil {
			writeInfraiError(w, err)
			return
		}
		if !valid {
			http.Error(w, "captcha rejected", http.StatusUnprocessableEntity)
			return
		}
		messageID, err := client.SendTeamEmail(r.Context(), mail, contactKey(form))
		if err != nil {
			writeInfraiError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message_id": messageID})
	}
}

func contactKey(form ContactSubmission) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(form.Email) + "\n" + strings.TrimSpace(form.Message)))
	return "creator-contact-" + hex.EncodeToString(sum[:])
}

func writeInfraiError(w http.ResponseWriter, err error) {
	var apiErr *InfraiError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
		http.Error(w, "submission rejected", apiErr.StatusCode)
		return
	}
	http.Error(w, "submission could not be processed", http.StatusBadGateway)
}
