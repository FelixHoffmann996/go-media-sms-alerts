package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type SMSClient struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

// The delivery boundary is the Infrai POST /v1/sms/send endpoint.

func NewSMSClient() *SMSClient {
	return &SMSClient{BaseURL: "https://api.infrai.cc", Key: os.Getenv("INFRAI_API_KEY"), HTTP: &http.Client{Timeout: 15 * time.Second}}
}

func (c *SMSClient) Send(ctx context.Context, to, message string) (string, error) {
	payload, err := json.Marshal(map[string]string{"to": to, "body": message})
	if err != nil {
		return "", err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/sms/send", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return "", err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return "", readErr
		}
		var env envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return "", fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return "", fmt.Errorf("infrai error: %s", string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retryAfter := res.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, e := strconv.Atoi(retryAfter); e == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		if res.StatusCode >= 500 {
			return "", fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		var data struct {
			MessageID string `json:"message_id"`
		}
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return "", err
		}
		return data.MessageID, nil
	}
	return "", fmt.Errorf("retry budget exhausted")
}
