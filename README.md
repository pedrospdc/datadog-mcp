# datadog-mcp

A Model Context Protocol (MCP) server that exposes Datadog's monitoring and observability data to AI models. This allows AI assistants like Claude to query metrics, traces, services, and dashboards from your Datadog account.

## Features

- **Metrics**: Query timeseries metrics and list available metrics
- **APM/Traces**: Query spans and get APM statistics (latency, error rates, throughput)
- **Service Catalog**: List services with metadata (team, tier, lifecycle, contacts)
- **Dashboards**: List and retrieve dashboard configurations

## Installation

```bash
go install github.com/pedrospdc/datadog-mcp/cmd/datadog-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/pedrospdc/datadog-mcp.git
cd datadog-mcp
make build
```

## Configuration

Set the following environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `DD_API_KEY` | Yes | Your Datadog API key |
| `DD_APP_KEY` | Yes | Your Datadog application key |
| `DD_SITE` | No | Datadog site (default: `datadoghq.com`) |

### Getting API Keys

1. Go to [Datadog API Keys](https://app.datadoghq.com/organization-settings/api-keys)
2. Create a new API key or use an existing one
3. Go to [Datadog Application Keys](https://app.datadoghq.com/organization-settings/application-keys)
4. Create a new application key

## Usage

### With Claude Code (CLI)

Claude Code can use MCP servers configured in your settings. Add the datadog-mcp server to your configuration:

**Option 1: Project-level configuration (recommended)**

Create or edit `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "datadog": {
      "command": "/path/to/datadog-mcp",
      "env": {
        "DD_API_KEY": "your-api-key",
        "DD_APP_KEY": "your-app-key"
      }
    }
  }
}
```

**Option 2: User-level configuration**

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "datadog": {
      "command": "/path/to/datadog-mcp",
      "env": {
        "DD_API_KEY": "your-api-key",
        "DD_APP_KEY": "your-app-key"
      }
    }
  }
}
```

**Option 3: Add via CLI**

```bash
claude mcp add datadog /path/to/datadog-mcp \
  -e DD_API_KEY=your-api-key \
  -e DD_APP_KEY=your-app-key
```

Once configured, you can ask Claude Code questions like:
- "What are the current CPU metrics for our production hosts?"
- "Show me the error rate for the payment-service over the last hour"
- "List all our Datadog dashboards"
- "What services do we have in the service catalog?"

### With Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "datadog": {
      "command": "/path/to/datadog-mcp",
      "env": {
        "DD_API_KEY": "your-api-key",
        "DD_APP_KEY": "your-app-key"
      }
    }
  }
}
```

### Standalone

```bash
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
./build/datadog-mcp
```

## Available Tools

### query_metrics

Query timeseries metrics data from Datadog.

**Parameters:**
- `query` (required): Datadog metric query string (e.g., `avg:system.cpu.user{*} by {host}`)
- `from`: Start time in RFC3339 format or relative (e.g., `now-1h`). Defaults to 1 hour ago
- `to`: End time in RFC3339 format or relative (e.g., `now`). Defaults to now

**Example:**
```
Query CPU usage across all hosts for the last hour
```

### list_metrics

List available metrics in Datadog.

**Parameters:**
- `tag_filter`: Filter metrics by tag (e.g., `env:production`)
- `host`: Filter metrics by host name
- `prefix`: Filter metrics by name prefix

**Example:**
```
List all metrics with prefix "aws.ec2"
```

### get_apm_services

List all APM services from the Datadog service catalog.

**Parameters:** None

**Returns:** Service metadata including team, tier, lifecycle, languages, and contacts.

### query_spans

Query APM spans/traces from Datadog.

**Parameters:**
- `query`: Span search query (e.g., `service:my-service` or `@http.status_code:500`). Defaults to `*`
- `from`: Start time (e.g., `now-15m`). Defaults to `now-15m`
- `to`: End time (e.g., `now`). Defaults to now
- `limit`: Maximum spans to return (1-1000). Defaults to 50

**Example:**
```
Find all spans with 500 errors in the payment service
```

### query_apm_stats

Query APM statistics for a service.

**Parameters:**
- `service` (required): Service name to query
- `operation`: Specific operation/resource name
- `env`: Environment filter (e.g., `production`)
- `from`: Start time. Defaults to 1 hour ago
- `to`: End time. Defaults to now

**Returns:** Latency percentiles (avg, p50, p95, p99), error rates, and throughput.

### list_dashboards

List all dashboards in your Datadog account.

**Parameters:**
- `filter_shared`: Filter to only shared dashboards
- `filter_deleted`: Include deleted dashboards

**Example:**
```
Show me all available dashboards
```

### get_dashboard

Get detailed information about a specific dashboard.

**Parameters:**
- `dashboard_id` (required): The dashboard ID to retrieve

**Returns:** Dashboard configuration including title, description, layout type, widgets, and template variables.

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Run locally
make run
```

## License

MIT
