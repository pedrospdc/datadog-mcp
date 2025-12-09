package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/datadog"
)

// RegisterAll registers all Datadog tools with the MCP server.
func RegisterAll(server *mcp.Server, client *datadog.Client) {
	registerQueryMetrics(server, client)
	registerListMetrics(server, client)
	registerGetAPMServices(server, client)
	registerQuerySpans(server, client)
	registerQueryAPMStats(server, client)
	registerListDashboards(server, client)
	registerGetDashboard(server, client)
}
