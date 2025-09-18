package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/openai/openai-go/v2"
)

// TodayTool provides the current date and time in RFC3339 format.
// This tool requires no parameters and returns the current timestamp.
type TodayTool struct{}

func (TodayTool) Name() string { return "get_today_date" }

func (TodayTool) Schema() openai.FunctionDefinitionParam {
	return openai.FunctionDefinitionParam{
		Name:        "get_today_date",
		Description: openai.String("Get today's date and time in RFC3339 format"),
	}
}

func (TodayTool) Handle(ctx context.Context, args json.RawMessage) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}
