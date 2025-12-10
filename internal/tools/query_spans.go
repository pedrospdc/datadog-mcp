package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// QuerySpansInput defines the input for the query_spans tool.
type QuerySpansInput struct {
	Query  string `json:"query,omitempty" jsonschema:"Span search query, e.g. service:my-service or @http.status_code:500. Defaults to * (all spans)"`
	From   string `json:"from,omitempty" jsonschema:"Start time, e.g. now-15m or now-1h. Defaults to now-15m"`
	To     string `json:"to,omitempty" jsonschema:"End time, e.g. now. Defaults to now"`
	Limit  int32  `json:"limit,omitempty" jsonschema:"Maximum number of spans to return (1-1000). Defaults to 50"`
	Cursor string `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response to get next page of results"`
}

func registerQuerySpans(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_spans",
		Description: "Query APM spans/traces from Datadog. Search for specific spans by service, operation, status code, or custom tags.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input QuerySpansInput) (*mcp.CallToolResult, *datadog.QuerySpansResult, error) {
		query := input.Query
		if query == "" {
			query = "*"
		}

		result, err := client.QuerySpans(ctx, query, input.From, input.To, input.Limit, input.Cursor)
		if err != nil {
			return nil, nil, err
		}

		summary := fmt.Sprintf("Found %d spans matching query: %s\n\n", result.TotalCount, query)

		for i, span := range result.Spans {
			if i >= 20 {
				summary += fmt.Sprintf("\n... and %d more spans (see structured output for full results)", result.TotalCount-20)
				break
			}
			summary += fmt.Sprintf("[%d] %s / %s\n", i+1, span.Service, span.Name)
			summary += fmt.Sprintf("    Resource: %s\n", span.Resource)
			summary += fmt.Sprintf("    Status: %s, Duration: %.2fms\n", span.Status, float64(span.Duration)/1e6)
			summary += fmt.Sprintf("    TraceID: %s, SpanID: %s\n", span.TraceID, span.SpanID)
			summary += "\n"
		}

		if result.NextCursor != "" {
			summary += fmt.Sprintf("\nNext page cursor: %s", result.NextCursor)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
