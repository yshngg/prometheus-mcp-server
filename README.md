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

| Flag | Default | Description |
|------|---------|-------------|
| `-prom-addr` | `http://localhost:9090` | Prometheus server URL |
| `-mcp-addr` | `localhost:8080` | MCP server listen address |
| `-transport` | `stdio` | `stdio`, `http`, or `sse` |
| `-version` | | Print version |

## MCP Client Configuration

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
      "args": ["run", "--rm", "-i", "ghcr.io/yshngg/prometheus-mcp-server:latest", "--prom-addr", "http://host.docker.internal:9090"]
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
      "args": ["run", "--rm", "-i", "ghcr.io/yshngg/prometheus-mcp-server:latest", "--prom-addr", "http://host.docker.internal:9090"]
    }
  }
}
```

</details>

<details>
  <summary>VS Code / Copilot</summary>

Add to MCP settings (`~/.vscode/mcp.json` or Settings → MCP):

```json
{
  "servers": {
    "prometheus": {
      "command": "prometheus-mcp-server",
      "args": ["--prom-addr", "http://localhost:9090"]
    }
  }
}
```

</details>

<details>
  <summary>Cline</summary>

Follow the <a href="https://docs.cline.bot/mcp/configuring-mcp-servers">Cline MCP guide</a> and use the standard config above.

</details>

<details>
  <summary>Cursor</summary>

Go to `Cursor Settings` → `MCP` → `New MCP Server`. Use the standard config above.

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
      "command": ["prometheus-mcp-server", "--prom-addr", "http://localhost:9090"]
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

</details>

## Tools

**Expression:** `instant-query`, `range-query`

**Metadata:** `find-series-by-labels`, `list-label-names`, `list-label-values`, `target-metadata-query`, `metric-metadata-query`

**Discovery:** `target-discovery`, `alert-query`, `rule-query`, `alertmanager-discovery`

**Status:** `config`, `flags`, `build-information`, `runtime-information`, `tsdb-stats`, `wal-replay-stats`

**TSDB Admin:** `tsdb-snapshot`, `delete-series`, `clean-tombstones`

**Management:** `health-check`, `readiness-check`, `reload`, `quit`

## License

Apache License 2.0
