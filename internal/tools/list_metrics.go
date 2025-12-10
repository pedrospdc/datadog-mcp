package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// ListMetricsInput defines the input for the list_metrics tool.
type ListMetricsInput struct {
	TagFilter string `json:"tag_filter,omitempty" jsonschema:"Filter metrics by tag, e.g. env:production"`
	Host      string `json:"host,omitempty" jsonschema:"Filter metrics by host name"`
	Prefix    string `json:"prefix,omitempty" jsonschema:"Filter metrics by name prefix (client-side filtering)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of metrics to return per page. Defaults to 100"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Number of metrics to skip for pagination. Defaults to 0"`
}

func registerListMetrics(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_metrics",
		Description: "List available metrics in Datadog. Can filter by tag, host, or name prefix.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListMetricsInput) (*mcp.CallToolResult, *datadog.ListMetricsResult, error) {
		from := time.Now().Add(-24 * time.Hour)

		result, err := client.ListMetrics(ctx, from, input.Host, input.TagFilter)
		if err != nil {
			return nil, nil, err
		}

		// Apply prefix filter client-side if specified
		if input.Prefix != "" {
			filtered := make([]string, 0)
			for _, metric := range result.Metrics {
				if strings.HasPrefix(metric, input.Prefix) {
					filtered = append(filtered, metric)
				}
			}
			result.Metrics = filtered
		}

		totalMetrics := len(result.Metrics)

		// Apply pagination (client-side since Datadog API doesn't support it for list metrics)
		limit := input.Limit
		if limit <= 0 {
			limit = 100
		}
		offset := input.Offset
		if offset < 0 {
			offset = 0
		}

		// Slice the metrics for pagination
		start := offset
		if start > len(result.Metrics) {
			start = len(result.Metrics)
		}
		end := start + limit
		if end > len(result.Metrics) {
			end = len(result.Metrics)
		}

		paginatedMetrics := result.Metrics[start:end]
		hasMore := end < len(result.Metrics)

		summary := fmt.Sprintf("Found %d metrics total", totalMetrics)
		if input.TagFilter != "" {
			summary += fmt.Sprintf(" (tag filter: %s)", input.TagFilter)
		}
		if input.Host != "" {
			summary += fmt.Sprintf(" (host: %s)", input.Host)
		}
		if input.Prefix != "" {
			summary += fmt.Sprintf(" (prefix: %s)", input.Prefix)
		}
		summary += fmt.Sprintf("\nShowing %d-%d of %d:\n\n", start+1, start+len(paginatedMetrics), totalMetrics)

		for _, metric := range paginatedMetrics {
			summary += metric + "\n"
		}

		if hasMore {
			summary += fmt.Sprintf("\nMore results available. Use offset=%d to get the next page.", end)
		}

		// Update result with paginated metrics for structured output
		result.Metrics = paginatedMetrics
		result.Total = totalMetrics
		result.Offset = offset
		result.HasMore = hasMore

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
