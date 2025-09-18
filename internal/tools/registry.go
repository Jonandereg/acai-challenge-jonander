package tools

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/openai/openai-go/v2"
)

type Tool interface {
	Name() string
	Schema() openai.FunctionDefinitionParam
	Handle(ctx context.Context, args json.RawMessage) (string, error)
}

type Registry struct {
	byName map[string]Tool
}

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
func (r *Registry) ToolsForOpen() []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(r.byName))
	for _, t := range r.byName {
		out = append(out, openai.ChatCompletionFunctionTool(t.Schema()))
	}
	return out
}

// Dispatch handles a single tool call (by name) and returns the tool output string.
func (r *Registry) Dispatch(ctx context.Context, name string, args json.RawMessage) (string, error) {
	t, ok := r.byName[name]
	if !ok {
		return "", errors.New("unknown tool: " + name)
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return t.Handle(cctx, args)
}

// Helper for arg parsing in each tool.
func ParseArgs[T any](raw json.RawMessage, out *T) error {
	return json.Unmarshal(raw, out)
}
