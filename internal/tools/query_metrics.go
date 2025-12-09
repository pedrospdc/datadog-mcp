package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// QueryMetricsInput defines the input for the query_metrics tool.
type QueryMetricsInput struct {
	Query string `json:"query" jsonschema:"required,description=Datadog metric query string (e.g. avg:system.cpu.user{*} by {host})"`
	From  string `json:"from" jsonschema:"description=Start time in RFC3339 format or relative (e.g. now-1h). Defaults to 1 hour ago"`
	To    string `json:"to" jsonschema:"description=End time in RFC3339 format or relative (e.g. now). Defaults to now"`
}

func registerQueryMetrics(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_metrics",
		Description: "Query timeseries metrics data from Datadog. Returns metric values over a time range with support for aggregations and grouping.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QueryMetricsInput) (*mcp.CallToolResult, *datadog.QueryMetricsResult, error) {
		var from, to time.Time
		var err error

		if input.From == "" {
			from = time.Now().Add(-1 * time.Hour)
		} else {
			from, err = parseTime(input.From)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'from' time: %w", err)
			}
		}

		if input.To == "" {
			to = time.Now()
		} else {
			to, err = parseTime(input.To)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid 'to' time: %w", err)
			}
		}

		if from.After(to) {
			return nil, nil, fmt.Errorf("'from' time must be before 'to' time")
		}

		result, err := client.QueryMetrics(ctx, input.Query, from, to)
		if err != nil {
			return nil, nil, err
		}

		summary := fmt.Sprintf("Query: %s\nTime Range: %s to %s\nSeries Count: %d\n",
			input.Query, from.Format(time.RFC3339), to.Format(time.RFC3339), len(result.Series))

		for i, series := range result.Series {
			summary += fmt.Sprintf("\n[%d] %s (%d data points)", i+1, series.Metric, len(series.DataPoints))
			if len(series.Tags) > 0 {
				summary += fmt.Sprintf(" - Tags: %v", series.Tags)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
