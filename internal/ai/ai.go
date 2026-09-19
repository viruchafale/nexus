// Package ai implements the optional `nexus ask` assistant.
//
// Architecture: CLI -> CollectContext -> Provider -> LLM -> response text.
//
// The AI never executes commands: it receives a read-only text snapshot
// and its reply is printed, nothing more. No secrets are ever hardcoded
// or printed; the API key travels only as an HTTPS bearer token.
package ai

import "context"

// Request is one completion call.
type Request struct {
	Model    string
	System   string
	Question string
	Context  string
}

// Provider is the seam behind which all provider-specific code lives.
// Only OpenAI-compatible backends exist today; add new ones by
// implementing this interface, never by branching in callers.
type Provider interface {
	Name() string
	Complete(ctx context.Context, req Request) (string, error)
}

// SystemPrompt instructs the model to stay advisory and read-only.
const SystemPrompt = `You are NEXUS, a concise DevOps troubleshooting assistant. ` +
	`You receive a read-only snapshot of the user's development machine plus a question. ` +
	`Rules: answer concisely with concrete next steps; suggest only safe, read-only ` +
	`diagnostics first; never claim to execute commands; never ask for secrets.`
