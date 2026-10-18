package main

import (
	"errors"
	"fmt"
	"html"
	"strings"
)

// ContactSubmission is the public form contract. Consent is recorded in the inbox copy.
type ContactSubmission struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	Message      string `json:"message"`
	CaptchaToken string `json:"captcha_token"`
	Consent      bool   `json:"consent"`
}

type TeamEmail struct {
	To      string
	Subject string
	HTML    string
}

func routeCreatorContact(in ContactSubmission, inbox string) (TeamEmail, error) {
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Message) == "" {
		return TeamEmail{}, errors.New("email and message are required")
	}
	if !in.Consent {
		return TeamEmail{}, errors.New("contact consent is required")
	}
	if strings.TrimSpace(inbox) == "" {
		return TeamEmail{}, errors.New("team inbox is required")
	}
	name := html.EscapeString(strings.TrimSpace(in.Name))
	message := html.EscapeString(strings.TrimSpace(in.Message))
	email := html.EscapeString(strings.TrimSpace(in.Email))
	return TeamEmail{
		To:      inbox,
		Subject: "Creator contact: " + strings.TrimSpace(in.Email),
		HTML:    fmt.Sprintf("<h1>Creator contact</h1><p><b>Name:</b> %s</p><p><b>Email:</b> %s</p><p><b>Consent:</b> recorded</p><p>%s</p>", name, email, message),
	}, nil
}
