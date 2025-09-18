package tools

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/openai/openai-go/v2"
)

// Tool defines the interface that all tools must implement to be registered
// and used within the system. Each tool provides its name, OpenAI function schema,
// and handles execution with the given arguments.
type Tool interface {
	Name() string
	Schema() openai.FunctionDefinitionParam
	Handle(ctx context.Context, args json.RawMessage) (string, error)
}

// Registry manages a collection of tools, providing registration, schema exposure,
// and dispatch functionality for tool execution.
type Registry struct {
	byName map[string]Tool
}

// NewRegistry creates a new tool registry with the provided tools.
// Each tool is indexed by its name for efficient lookup during dispatch.
func NewRegistry(ts ...Tool) *Registry {
	m := make(map[string]Tool, len(ts))
	for _, t := range ts {
		m[t.Name()] = t
	}
	return &Registry{
		byName: m,
	}
}

// ToolsForOpenAI exposes the JSON Function schemas to the model.
// Returns all registered tools formatted as OpenAI function definitions.
func (r *Registry) ToolsForOpenAI() []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(r.byName))
	for _, t := range r.byName {
		out = append(out, openai.ChatCompletionFunctionTool(t.Schema()))
	}
	return out
}

// Dispatch handles a single tool call (by name) and returns the tool output string.
// Executes the named tool with the provided arguments within a 5-second timeout.
func (r *Registry) Dispatch(ctx context.Context, name string, args json.RawMessage) (string, error) {
	t, ok := r.byName[name]
	if !ok {
		return "", errors.New("unknown tool: " + name)
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return t.Handle(cctx, args)
}

// ParseArgs is a helper function for parsing JSON arguments in tool implementations.
// It unmarshals the raw JSON message into the provided output structure.
func ParseArgs[T any](raw json.RawMessage, out *T) error {
	return json.Unmarshal(raw, out)
}
