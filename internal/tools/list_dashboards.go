package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// ListDashboardsInput defines the input for the list_dashboards tool.
type ListDashboardsInput struct {
	FilterShared  bool `json:"filter_shared,omitempty" jsonschema:"Filter to only shared dashboards"`
	FilterDeleted bool `json:"filter_deleted,omitempty" jsonschema:"Include deleted dashboards"`
}

func registerListDashboards(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dashboards",
		Description: "List all dashboards in your Datadog account. Returns dashboard titles, IDs, layout types, and metadata.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDashboardsInput) (*mcp.CallToolResult, *datadog.ListDashboardsResult, error) {
		result, err := client.ListDashboards(ctx, input.FilterShared, input.FilterDeleted)
		if err != nil {
			return nil, nil, err
		}

		summary := fmt.Sprintf("Found %d dashboards:\n\n", result.Total)

		for i, d := range result.Dashboards {
			if i >= 50 {
				summary += fmt.Sprintf("\n... and %d more dashboards", result.Total-50)
				break
			}
			summary += fmt.Sprintf("[%s] %s\n", d.ID, d.Title)
			if d.Description != "" {
				summary += fmt.Sprintf("  Description: %s\n", d.Description)
			}
			summary += fmt.Sprintf("  Layout: %s\n", d.LayoutType)
			if d.AuthorHandle != "" {
				summary += fmt.Sprintf("  Author: %s\n", d.AuthorHandle)
			}
			if !d.ModifiedAt.IsZero() {
				summary += fmt.Sprintf("  Modified: %s\n", d.ModifiedAt.Format("2006-01-02 15:04:05"))
			}
			summary += "\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
