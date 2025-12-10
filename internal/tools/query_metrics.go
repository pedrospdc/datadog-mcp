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
	Query         string `json:"query" jsonschema:"Datadog metric query string, e.g. avg:system.cpu.user{*} by {host}"`
	From          string `json:"from,omitempty" jsonschema:"Start time in RFC3339 format or relative, e.g. now-1h. Defaults to 1 hour ago"`
	To            string `json:"to,omitempty" jsonschema:"End time in RFC3339 format or relative, e.g. now. Defaults to now"`
	MaxDataPoints int    `json:"max_data_points,omitempty" jsonschema:"Maximum number of data points to return per series. Defaults to 300. Use 0 for unlimited."`
	MaxSeries     int    `json:"max_series,omitempty" jsonschema:"Maximum number of series to return. Defaults to 100. Use 0 for unlimited."`
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

		// Apply pagination limits
		maxSeries := input.MaxSeries
		if maxSeries <= 0 {
			maxSeries = 100
		}
		maxDataPoints := input.MaxDataPoints
		if maxDataPoints <= 0 {
			maxDataPoints = 300
		}

		totalSeries := len(result.Series)
		truncatedSeries := false
		truncatedDataPoints := false

		// Limit number of series
		if maxSeries > 0 && len(result.Series) > maxSeries {
			result.Series = result.Series[:maxSeries]
			truncatedSeries = true
		}

		// Limit data points per series
		for i := range result.Series {
			if maxDataPoints > 0 && len(result.Series[i].DataPoints) > maxDataPoints {
				result.Series[i].DataPoints = result.Series[i].DataPoints[:maxDataPoints]
				truncatedDataPoints = true
			}
		}

		summary := fmt.Sprintf("Query: %s\nTime Range: %s to %s\nSeries Count: %d",
			input.Query, from.Format(time.RFC3339), to.Format(time.RFC3339), len(result.Series))

		if truncatedSeries {
			summary += fmt.Sprintf(" (truncated from %d, use max_series to see more)", totalSeries)
		}
		summary += "\n"

		for i, series := range result.Series {
			dataPointInfo := fmt.Sprintf("%d data points", len(series.DataPoints))
			if truncatedDataPoints {
				dataPointInfo += " (truncated, use max_data_points to see more)"
			}
			summary += fmt.Sprintf("\n[%d] %s (%s)", i+1, series.Metric, dataPointInfo)
			if len(series.Tags) > 0 {
				summary += fmt.Sprintf(" - Tags: %v", series.Tags)
			}
		}

		// Add pagination info to result
		result.TotalSeries = totalSeries
		result.Truncated = truncatedSeries || truncatedDataPoints

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
