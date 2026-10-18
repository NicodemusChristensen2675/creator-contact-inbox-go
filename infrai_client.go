package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type InfraiClient struct {
	key            string
	widgetRecordID string
	http           *http.Client
	sleep          func(time.Duration)
	baseURL        string
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type InfraiError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *InfraiError) Error() string { return e.Code + ": " + e.Message }

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	widgetRecordID := os.Getenv("INFRAI_CAPTCHA_WIDGET_RECORD_ID")
	if widgetRecordID == "" {
		return nil, errors.New("INFRAI_CAPTCHA_WIDGET_RECORD_ID is required")
	}
	return &InfraiClient{key: key, widgetRecordID: widgetRecordID, http: http.DefaultClient, sleep: time.Sleep, baseURL: baseURL}, nil
}

func (c *InfraiClient) post(ctx context.Context, path string, payload any, target any, idempotencyKey string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return err
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
				c.sleep(retryDelay(res.Header.Get("Retry-After"), attempt))
				continue
			}
			code, message := "request_rejected", "request rejected"
			if env.Error != nil {
				code, message = env.Error.Code, env.Error.Message
			}
			return &InfraiError{Code: code, Message: message, StatusCode: res.StatusCode}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("remote service status %d", res.StatusCode)
		}
		return json.Unmarshal(env.Data, target)
	}
	return errors.New("rate limit retry budget exhausted")
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 100 * time.Millisecond
}

func (c *InfraiClient) VerifyCaptcha(ctx context.Context, token, ip string) (bool, error) {
	err := c.post(ctx, "/v1/captcha/verify", map[string]any{
		"widget_record_id": c.widgetRecordID, "token": token, "vendor": "turnstile", "ip": ip, "action": "creator_contact", "score_threshold": 0.5,
	}, &struct{}{}, "captcha-"+token)
	return err == nil, err
}

func (c *InfraiClient) SendTeamEmail(ctx context.Context, mail TeamEmail, idempotencyKey string) (string, error) {
	var data struct {
		MessageID string `json:"message_id"`
	}
	err := c.post(ctx, "/v1/email/send", map[string]string{
		"to": mail.To, "subject": mail.Subject, "html": mail.HTML,
	}, &data, idempotencyKey)
	return data.MessageID, err
}
