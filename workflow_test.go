package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestAndNotifyMarksProcessedAndDelivers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/sms/send" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["to"] != "+15550001" || body["body"] != "Asset Trailer is ready" {
			t.Fatalf("wrong recipient: %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]string{"message_id": "msg-7"}})
	}))
	defer server.Close()
	p := NewPipeline(&SMSClient{BaseURL: server.URL, Key: "test", HTTP: server.Client()})
	job, delivery, err := p.IngestAndNotify(context.Background(), Asset{ID: "asset-1", Title: "Trailer"}, "Mina", "+15550001")
	if err != nil || job.State != "processed" || delivery.MessageID != "msg-7" {
		t.Fatalf("unexpected result: job=%+v delivery=%+v err=%v", job, delivery, err)
	}
}
