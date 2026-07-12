# Prometheus MCP Server

[![codecov](https://codecov.io/gh/yshngg/prometheus-mcp-server/graph/badge.svg?token=C64XY9GFP3)](https://codecov.io/gh/yshngg/prometheus-mcp-server)

A Go-based MCP server that exposes Prometheus query capabilities via the Model Context Protocol.

## Installation

```bash
go install github.com/yshngg/prometheus-mcp-server@latest
```

Or pull the Docker image:

```bash
docker pull ghcr.io/yshngg/prometheus-mcp-server:latest
```

## Usage

```bash
prometheus-mcp-server --prom-addr="http://localhost:9090"
```

### Flags

| Flag          | Default                 | Description                                             |
| ------------- | ----------------------- | ------------------------------------------------------- |
| `-prom-addr`  | `http://localhost:9090` | Prometheus server URL                                   |
| `-mcp-addr`   | `localhost:8080`        | MCP server listen address                               |
| `-transport`  | `stdio`                 | `stdio` or `http`                                       |
| `-auth-token` | ``                      | Bearer token for MCP endpoint authentication (optional) |
| `-version`    |                         | Print version                                           |

### Environment Variables

All flags can be set via environment variables: `PROM_ADDR`, `MCP_ADDR`, `TRANSPORT`, `AUTH_TOKEN`.

### MCP Client Configuration

Add the following config to your MCP client:

```json
{
  "mcpServers": {
    "prometheus": {
      "command": "prometheus-mcp-server",
      "args": ["--prom-addr", "http://localhost:9090"]
    }
  }
}
```

> [!NOTE]
> Install the binary via `go install github.com/yshngg/prometheus-mcp-server@latest` or pull the Docker image with `docker pull ghcr.io/yshngg/prometheus-mcp-server:latest`.

If you don't have the binary installed, use Docker:

```json
{
  "mcpServers": {
    "prometheus": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "ghcr.io/yshngg/prometheus-mcp-server:latest",
        "--prom-addr",
        "http://host.docker.internal:9090"
      ]
    }
  }
}
```

### Client-specific configuration

<details>
  <summary>Claude Desktop</summary>

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "prometheus": {
      "command": "prometheus-mcp-server",
      "args": ["--prom-addr", "http://localhost:9090"]
    }
  }
}
```

With Docker:

```json
{
  "mcpServers": {
    "prometheus": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "ghcr.io/yshngg/prometheus-mcp-server:latest",
        "--prom-addr",
        "http://host.docker.internal:9090"
      ]
    }
  }
}
```

</details>

<details>
  <summary>OpenCode</summary>

Add to `opencode.json` (<a href="https://opencode.ai/docs/mcp-servers">guide</a>):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "prometheus": {
      "type": "local",
      "command": [
        "prometheus-mcp-server",
        "--prom-addr",
        "http://localhost:9090"
      ]
    }
  }
}
```

</details>

<details>
  <summary>HTTP Transport</summary>

Run the server with HTTP transport:

```bash
prometheus-mcp-server --transport=http --mcp-addr="localhost:8080"
```

Then configure your client to use `http://localhost:8080/mcp`.

For authenticated access, set `AUTH_TOKEN`:

```bash
AUTH_TOKEN=my-token prometheus-mcp-server --transport=http
```

Your client must then include `Authorization: Bearer my-token` in requests to `/mcp`.

</details>

## Tools

**Expression:** `instant_query`, `range_query`

**Metadata:** `find_series_by_labels`, `list_label_names`, `list_label_values`, `target_metadata_query`, `metric_metadata_query`

**Discovery:** `target_discovery`, `alert_query`, `rule_query`, `alertmanager_discovery`

**TSDB Admin:** `tsdb_snapshot`, `delete_series`, `clean_tombstones`

**Management:** `health_check`, `readiness_check`, `reload`, `quit`

## Resources

The server exposes Prometheus data as URI-addressable resources under the `prom:///` scheme:

| URI                        | Description                           |
| -------------------------- | ------------------------------------- |
| `prom:///config`           | Currently loaded configuration (YAML) |
| `prom:///flags`            | Command-line flag values              |
| `prom:///runtime-info`     | Runtime information                   |
| `prom:///build-info`       | Build information                     |
| `prom:///tsdb-stats`       | TSDB cardinality statistics           |
| `prom:///wal-replay-stats` | WAL replay state                      |

### Resource Templates

URI templates for dynamic data access:

| Template                              | Description                         |
| ------------------------------------- | ----------------------------------- |
| `prom:///api/v1/query?query={promql}` | Instant PromQL query result         |
| `prom:///api/v1/label/{name}/values`  | Label values for a given label name |

## Prompts

| Name                    | Arguments           | Description                                       |
| ----------------------- | ------------------- | ------------------------------------------------- |
| `all-available-metrics` | `prefix` (optional) | Lists all metric names in the Prometheus instance |

## License

Apache License 2.0
