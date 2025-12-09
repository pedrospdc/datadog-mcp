package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// GetDashboardInput defines the input for the get_dashboard tool.
type GetDashboardInput struct {
	DashboardID string `json:"dashboard_id" jsonschema:"required,description=The dashboard ID to retrieve"`
}

func registerGetDashboard(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Get detailed information about a specific dashboard including its configuration, widgets, and template variables.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetDashboardInput) (*mcp.CallToolResult, *datadog.Dashboard, error) {
		result, err := client.GetDashboard(ctx, input.DashboardID)
		if err != nil {
			return nil, nil, err
		}

		summary := fmt.Sprintf("Dashboard: %s\n", result.Title)
		summary += fmt.Sprintf("ID: %s\n", result.ID)
		if result.Description != "" {
			summary += fmt.Sprintf("Description: %s\n", result.Description)
		}
		summary += fmt.Sprintf("Layout Type: %s\n", result.LayoutType)
		if result.URL != "" {
			summary += fmt.Sprintf("URL: %s\n", result.URL)
		}
		if result.AuthorHandle != "" {
			summary += fmt.Sprintf("Author: %s", result.AuthorHandle)
			if result.AuthorName != "" {
				summary += fmt.Sprintf(" (%s)", result.AuthorName)
			}
			summary += "\n"
		}
		if !result.CreatedAt.IsZero() {
			summary += fmt.Sprintf("Created: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		if !result.ModifiedAt.IsZero() {
			summary += fmt.Sprintf("Modified: %s\n", result.ModifiedAt.Format("2006-01-02 15:04:05"))
		}
		if result.IsReadOnly {
			summary += "Read Only: Yes\n"
		}
		if len(result.Tags) > 0 {
			summary += fmt.Sprintf("Tags: %v\n", result.Tags)
		}

		summary += fmt.Sprintf("\nWidgets: %d\n", result.WidgetCount)

		if len(result.TemplateVariables) > 0 {
			summary += "\nTemplate Variables:\n"
			for _, tv := range result.TemplateVariables {
				summary += fmt.Sprintf("  - %s", tv.Name)
				if tv.Prefix != "" {
					summary += fmt.Sprintf(" (prefix: %s)", tv.Prefix)
				}
				if tv.Default != "" {
					summary += fmt.Sprintf(" [default: %s]", tv.Default)
				}
				summary += "\n"
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: summary},
			},
		}, result, nil
	})
}
