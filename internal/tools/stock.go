package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/openai/openai-go/v2"
)

type stockResponse struct {
	Current float64 `json:"c"`
	High    float64 `json:"h"`
	Low     float64 `json:"l"`
	Open    float64 `json:"o"`
	Prev    float64 `json:"pc"`
}

// StockTool provides real-time stock market quotes for given ticker symbols
// using the Finnhub API service.
type StockTool struct{}

func (StockTool) Name() string {
	return "get_stock_quote"
}

func (StockTool) Schema() openai.FunctionDefinitionParam {
	return openai.FunctionDefinitionParam{
		Name:        "get_stock_quote",
		Description: openai.String("Get the current market value for a given stock symbol"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"symbol": map[string]string{
					"type":        "string",
					"description": "Ticker symbol, e.g. AAPL, TSLA, MSFT",
				},
			},
			"required": []string{"symbol"},
		},
	}
}

func (StockTool) Handle(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Symbol string `json:"symbol"`
	}
	if err := ParseArgs(args, &p); err != nil || strings.TrimSpace(p.Symbol) == "" {
		return "", errors.New("could not parse symbol")
	}

	data, err := fetchStock(ctx, p.Symbol)
	if err != nil {
		return "", errors.New("stock service unavailable")
	}

	return fmt.Sprintf("Current price for %s: $%.2f (high $%.2f, low $%.2f, open $%.2f, prev close $%.2f)",
		strings.ToUpper(p.Symbol), data.Current, data.High, data.Low, data.Open, data.Prev), nil
}

func fetchStock(ctx context.Context, symbol string) (*stockResponse, error) {
	token := os.Getenv("FINNHUB_TOKEN")
	if token == "" {
		return nil, errors.New("FINNHUB_TOKEN not set")
	}
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
		url.QueryEscape(symbol), token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("finnhub returned %s", res.Status)
	}

	var data stockResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
