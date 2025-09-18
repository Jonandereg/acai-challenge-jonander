package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/openai/openai-go/v2"
)

// CalendarTool provides functionality to retrieve local bank and public holidays
// from an ICS calendar feed. It supports filtering by date ranges and limiting results.
type CalendarTool struct{}

func (CalendarTool) Name() string { return "get_holidays" }
func (CalendarTool) Schema() openai.FunctionDefinitionParam {
	return openai.FunctionDefinitionParam{
		Name:        "get_holidays",
		Description: openai.String("Gets local bank and public holidays. Each line is a single holiday in the format 'YYYY-MM-DD: Holiday Name'."),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"before_date": map[string]string{
					"type":        "string",
					"description": "Optional date in RFC3339 format to get holidays before this date. If not provided, all holidays will be returned.",
				},
				"after_date": map[string]string{
					"type":        "string",
					"description": "Optional date in RFC3339 format to get holidays after this date. If not provided, all holidays will be returned.",
				},
				"max_count": map[string]string{
					"type":        "integer",
					"description": "Optional maximum number of holidays to return. If not provided, all holidays will be returned.",
				},
			},
		},
	}
}

func (CalendarTool) Handle(ctx context.Context, args json.RawMessage) (string, error) {
	link := "https://www.officeholidays.com/ics/spain/catalonia"
	if v := os.Getenv("HOLIDAY_CALENDAR_LINK"); v != "" {
		link = v
	}

	events, err := loadCalendar(ctx, link)
	if err != nil {
		return "", errors.New("failed to load holiday events")

	}
	var p struct {
		BeforeDate time.Time `json:"before_date,omitempty"`
		AfterDate  time.Time `json:"after_date,omitempty"`
		MaxCount   int       `json:"max_count,omitempty"`
	}
	if err := ParseArgs(args, &p); err != nil {
		return "", errors.New("failed to parse tool call arguments: " + err.Error())
	}

	var holidays []string
	for _, event := range events {
		date, err := event.GetAllDayStartAt()
		if err != nil {
			continue
		}

		if p.MaxCount > 0 && len(holidays) >= p.MaxCount {
			break
		}

		if !p.BeforeDate.IsZero() && date.After(p.BeforeDate) {
			continue
		}

		if !p.AfterDate.IsZero() && date.Before(p.AfterDate) {
			continue
		}

		holidays = append(holidays, date.Format(time.DateOnly)+": "+event.GetProperty(ics.ComponentPropertySummary).Value)
	}
	return strings.Join(holidays, "\n"), nil
}

func loadCalendar(ctx context.Context, link string) ([]*ics.VEvent, error) {
	slog.InfoContext(ctx, "Loading calendar", "link", link)

	cal, err := ics.ParseCalendarFromUrl(link, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	return cal.Events(), nil
}
