package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

type WebhookService struct {
	Client *http.Client
}

func NewWebhookService() *WebhookService {
	return &WebhookService{
		Client: &http.Client{
			// 3 minutes: covers slow AI workflows (LLM inference, tool calls, etc.)
			Timeout: 180 * time.Second,
		},
	}
}

type WebhookPayload struct {
	SessionID     string        `json:"session_id"`
	From          string        `json:"from"`
	SenderLID     string        `json:"sender_lid"`
	To            string        `json:"to"`
	Message       string        `json:"message"`
	Timestamp     time.Time     `json:"timestamp"`
	IsGroup       bool          `json:"is_group"`
	GroupInfo     *GroupInfo    `json:"group_info,omitempty"`
	PushName      string        `json:"push_name"`
	MessageType   string        `json:"message_type"`
	MediaData     []byte        `json:"-"` // Binary data, not for JSON
	MediaName     string        `json:"-"`
	MediaMimeType string        `json:"-"`
	Location      *LocationInfo `json:"location,omitempty"`
}

type GroupInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`    // Location name/address if available
	URL       string  `json:"url,omitempty"`     // URL to location if available
	IsLive    bool    `json:"is_live,omitempty"` // Whether it's a live location
}

func (s *WebhookService) SendWebhook(webhookURL string, payload WebhookPayload) (string, error) {
	if webhookURL == "" {
		return "", nil
	}

	// buildRequest creates a fresh *http.Request every call so retries get a new body reader.
	buildRequest := func() (*http.Request, error) {
		if len(payload.MediaData) > 0 {
			// Multipart/form-data for media payloads
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("session_id", payload.SessionID)
			_ = writer.WriteField("from", payload.From)
			_ = writer.WriteField("sender_lid", payload.SenderLID)
			_ = writer.WriteField("to", payload.To)
			_ = writer.WriteField("message", payload.Message)
			_ = writer.WriteField("timestamp", payload.Timestamp.Format(time.RFC3339))
			_ = writer.WriteField("is_group", fmt.Sprintf("%v", payload.IsGroup))
			_ = writer.WriteField("push_name", payload.PushName)
			_ = writer.WriteField("message_type", payload.MessageType)
			if payload.GroupInfo != nil {
				groupInfoJSON, _ := json.Marshal(payload.GroupInfo)
				_ = writer.WriteField("group_info", string(groupInfoJSON))
			}
			if payload.Location != nil {
				locationJSON, _ := json.Marshal(payload.Location)
				_ = writer.WriteField("location", string(locationJSON))
			}

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, payload.MediaName))
			h.Set("Content-Type", payload.MediaMimeType)
			part, _ := writer.CreatePart(h)
			part.Write(payload.MediaData)
			writer.Close()

			req, err := http.NewRequest("POST", webhookURL, body)
			if err != nil {
				return nil, fmt.Errorf("failed to create multipart request: %w", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			fmt.Printf("[Webhook] Built multipart request. Size: %d bytes\n", body.Len())
			return req, nil
		}

		// JSON payload (no media)
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal webhook payload: %w", err)
		}
		req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}

	// Retry up to 3 times, rebuilding the request each time to avoid consumed-body issues.
	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 2 * time.Second
			fmt.Printf("[Webhook] Retry %d/%d after %s...\n", attempt, maxAttempts-1, backoff)
			time.Sleep(backoff)
		}

		req, err := buildRequest()
		if err != nil {
			return "", err // build errors are not retryable
		}

		fmt.Printf("[Webhook] Sending request (attempt %d) to %s\n", attempt+1, webhookURL)
		resp, err := s.Client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			fmt.Printf("[Webhook] Attempt %d error: %v\n", attempt+1, err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("[Webhook] Response status: %d, body: %s\n", resp.StatusCode, string(bodyBytes))

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var data interface{}
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				// Not JSON — treat raw body as the response text
				return string(bodyBytes), nil
			}
			return extractText(data), nil
		}

		lastErr = fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return "", fmt.Errorf("webhook failed after %d attempts: %w", maxAttempts, lastErr)
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
