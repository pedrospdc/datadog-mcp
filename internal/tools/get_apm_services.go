package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// GetAPMServicesInput defines the input for the get_apm_services tool.
type GetAPMServicesInput struct {
	// No required inputs - lists all services
}

func registerGetAPMServices(server *mcp.Server, client *datadog.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_apm_services",
		Description: "List all APM services from the Datadog service catalog with their metadata including team, tier, lifecycle, and contacts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetAPMServicesInput) (*mcp.CallToolResult, *datadog.ListServicesResult, error) {
		result, err := client.ListServices(ctx)
		if err != nil {
			return nil, nil, err
		}

		summary := fmt.Sprintf("Found %d services:\n\n", result.Total)

		for _, svc := range result.Services {
			summary += fmt.Sprintf("Service: %s\n", svc.Name)
			if svc.Description != "" {
				summary += fmt.Sprintf("  Description: %s\n", svc.Description)
			}
			if svc.Team != "" {
				summary += fmt.Sprintf("  Team: %s\n", svc.Team)
			}
			if svc.Tier != "" {
				summary += fmt.Sprintf("  Tier: %s\n", svc.Tier)
			}
			if svc.Lifecycle != "" {
				summary += fmt.Sprintf("  Lifecycle: %s\n", svc.Lifecycle)
			}
			if len(svc.Languages) > 0 {
				summary += fmt.Sprintf("  Languages: %v\n", svc.Languages)
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
