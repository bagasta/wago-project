package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebhookService struct {
	Client *http.Client
}

func NewWebhookService() *WebhookService {
	return &WebhookService{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type WebhookPayload struct {
	SessionID   string     `json:"session_id"`
	From        string     `json:"from"`
	To          string     `json:"to"`
	Message     string     `json:"message"`
	Timestamp   time.Time  `json:"timestamp"`
	IsGroup     bool       `json:"is_group"`
	GroupInfo   *GroupInfo `json:"group_info,omitempty"`
	PushName    string     `json:"push_name"`
	MessageType string     `json:"message_type"`
}

type GroupInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *WebhookService) SendWebhook(webhookURL string, payload WebhookPayload) (string, error) {
	if webhookURL == "" {
		return "", nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Simple retry logic (3 times)
	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := s.Client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Read response body
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("[Webhook] Raw Response: %s\n", string(bodyBytes))

			var data interface{}
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				fmt.Printf("[Webhook] JSON Decode Error: %v\n", err)
				return "", nil
			}

			return extractText(data), nil
		}

		lastErr = fmt.Errorf("webhook returned status: %d", resp.StatusCode)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return "", fmt.Errorf("failed to send webhook after retries: %w", lastErr)
}

func extractText(data interface{}) string {
	switch v := data.(type) {
	case []interface{}:
		if len(v) > 0 {
			return extractText(v[0])
		}
	case map[string]interface{}:
		// Check common keys
		for _, key := range []string{"output", "text", "message", "response", "body", "content"} {
			if val, ok := v[key].(string); ok && val != "" {
				return val
			}
		}
		// Special case for nested "data" or "json"
		if val, ok := v["data"]; ok {
			return extractText(val)
		}
		if val, ok := v["json"]; ok {
			return extractText(val)
		}
	case string:
		return v
	}
	return ""
}
