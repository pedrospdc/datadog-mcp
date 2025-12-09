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
	TagFilter string `json:"tag_filter" jsonschema:"description=Filter metrics by tag (e.g. env:production)"`
	Host      string `json:"host" jsonschema:"description=Filter metrics by host name"`
	Prefix    string `json:"prefix" jsonschema:"description=Filter metrics by name prefix (client-side filtering)"`
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

		summary := fmt.Sprintf("Found %d metrics", len(result.Metrics))
		if input.TagFilter != "" {
			summary += fmt.Sprintf(" (tag filter: %s)", input.TagFilter)
		}
		if input.Host != "" {
			summary += fmt.Sprintf(" (host: %s)", input.Host)
		}
		if input.Prefix != "" {
			summary += fmt.Sprintf(" (prefix: %s)", input.Prefix)
		}
		summary += ":\n\n"

		maxDisplay := 100
		for i, metric := range result.Metrics {
			if i >= maxDisplay {
				summary += fmt.Sprintf("\n... and %d more metrics", len(result.Metrics)-maxDisplay)
				break
			}
			summary += metric + "\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
