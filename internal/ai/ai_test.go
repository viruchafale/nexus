package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeOpenAI serves a canned chat-completions response and records the request.
func fakeOpenAI(t *testing.T, status int, body string, check func(r *http.Request, payload map[string]any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("bad auth header %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("bad request JSON: %v", err)
		}
		if check != nil {
			check(r, payload)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestCompleteSuccess(t *testing.T) {
	srv := fakeOpenAI(t, 200, `{"choices":[{"message":{"content":"  Restart nothing; check logs.  "}}]}`,
		func(_ *http.Request, payload map[string]any) {
			if payload["model"] != "test-model" {
				t.Errorf("model = %v", payload["model"])
			}
			msgs, _ := payload["messages"].([]any)
			if len(msgs) != 2 {
				t.Errorf("want system+user messages, got %v", payload["messages"])
			}
		})
	defer srv.Close()

	p := &OpenAIProvider{APIKey: "test-key", BaseURL: srv.URL, HTTPClient: srv.Client()}
	got, err := p.Complete(context.Background(), Request{
		Model: "test-model", System: SystemPrompt, Question: "q?", Context: "ctx",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Restart nothing; check logs." {
		t.Errorf("response not trimmed/parsed: %q", got)
	}
}

func TestCompleteAPIError(t *testing.T) {
	srv := fakeOpenAI(t, 401, `{"error":{"message":"Incorrect API key"}}`, nil)
	defer srv.Close()

	p := &OpenAIProvider{APIKey: "test-key", BaseURL: srv.URL, HTTPClient: srv.Client()}
	_, err := p.Complete(context.Background(), Request{Model: "m", Question: "q"})
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Incorrect API key") {
		t.Errorf("want surfaced 401 + message, got %v", err)
	}
	if strings.Contains(err.Error(), "test-key") {
		t.Errorf("error must never contain the key: %v", err)
	}
}

func TestCompleteEmptyChoices(t *testing.T) {
	srv := fakeOpenAI(t, 200, `{"choices":[]}`, nil)
	defer srv.Close()

	p := &OpenAIProvider{APIKey: "test-key", BaseURL: srv.URL, HTTPClient: srv.Client()}
	if _, err := p.Complete(context.Background(), Request{Model: "m"}); err == nil {
		t.Error("empty choices should error")
	}
}

func TestCompleteMissingKey(t *testing.T) {
	p := &OpenAIProvider{BaseURL: "http://example.com"}
	if _, err := p.Complete(context.Background(), Request{}); err == nil {
		t.Error("missing key should error before any network call")
	}
}

func TestCompleteTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	p := &OpenAIProvider{APIKey: "k", BaseURL: srv.URL, HTTPClient: srv.Client()}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := p.Complete(ctx, Request{Model: "m"}); err == nil {
		t.Error("expired context should error")
	}
}

func TestCollectContextSections(t *testing.T) {
	ctx := CollectContext()
	for _, want := range []string{"OS:", "CPU:", "Memory:", "Disk", "Docker:", "Git:", "Network:", "Doctor:"} {
		if !strings.Contains(ctx, want) {
			t.Errorf("context missing %q\n%s", want, ctx)
		}
	}
}

func TestProviderName(t *testing.T) {
	var p Provider = &OpenAIProvider{}
	if p.Name() != "openai" {
		t.Errorf("got %q", p.Name())
	}
}
