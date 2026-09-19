# Phase 9 Report — Optional AI Assistant

Date: 2026-09-20
Goal: Optional `nexus ask "<question>"` (CLI → context collector → provider →
LLM → printed answer). Works fully without AI, no hardcoded/printed secrets,
provider behind an interface, timeouts, no command execution. No new Go deps.

## Files created/changed
- `internal/ai/ai.go` (new) — `Provider` interface + `Request`, read-only
  `SystemPrompt` (advisory only, never claims to execute, never asks secrets)
- `internal/ai/context.go` (new) — `CollectContext()` reusing system/docker/git/
  network/doctor (machine facts only: no env values, keys, or file contents)
- `internal/ai/openai.go` (new) — `OpenAIProvider` (stdlib `net/http` POST to
  `{base}/chat/completions`, bearer auth, key never in errors, 1MB body cap)
- `internal/ai/ai_test.go` (new) — httptest round-trip (path/auth/model/message
  shape, trim), 401 surfacing + key-absence-in-error, empty choices, missing
  key (no network), context timeout, context section headers, Name()
- `config/config.go` — added `AIProvider` (default `openai`), `AIBaseURL`
  (default api.openai.com/v1), `AITimeout` (`NEXUS_AI_TIMEOUT_SECS`, default
  60s); existing fields untouched
- `cmd/commands.go` — `askCmd`: no key → spec unavailable message, exit 0;
  unknown provider → explicit error; `--show-context` flag; timeout ctx;
  answer printed as text, never acted on
- `README.md` — full env table + "AI assistant (optional)" section

## Dependencies
- None new (stdlib `net/http`, `encoding/json` only)

## Commands to test
```bash
go test ./...
go run . ask "Why might my Docker backend be failing?"   # no key → setup msg
NEXUS_OPENAI_API_KEY=x NEXUS_AI_PROVIDER=foo go run . ask "hi"  # provider error
NEXUS_OPENAI_API_KEY=<key> go run . ask --show-context "Analyze my system."
go vet ./...
```

## Tests
- `internal/ai`: 7 tests incl. fake-server round-trip asserting endpoint path,
  bearer header, model passthrough, and system+user message shape
- Full suite: `go test ./...` ok (all 10 test packages)
- Manual: no-key path prints spec message exit 0; bad provider exits 1;
  end-to-end against a local stub server shows real context (OS/CPU/docker/
  git/network/doctor) + model reply, exit 0, key never in output

## Known issues
- Only OpenAI-compatible chat-completions backends; streaming not supported
  (single blocking call within the configured timeout)
- Context is a fixed section set; no `--no-context`/custom sections yet
- No conversation history — each `ask` is stateless
