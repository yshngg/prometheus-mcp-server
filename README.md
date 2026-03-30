# Prometheus Model Context Protocol Server

[![codecov](https://codecov.io/gh/yshngg/prometheus-mcp-server/graph/badge.svg?token=C64XY9GFP3)](https://codecov.io/gh/yshngg/prometheus-mcp-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/yshngg/prometheus-mcp-server)](https://goreportcard.com/report/github.com/yshngg/prometheus-mcp-server)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fyshngg%2Fprometheus-mcp-server.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fyshngg%2Fprometheus-mcp-server?ref=badge_shield)

**A Golang-based Model Context Protocol (MCP) server implementation for Prometheus that enables natural language interactions with Prometheus metrics and queries.**

**Built with Go**, `prometheus-mcp-server` provides a robust, type-safe interface that maintains full consistency with the Prometheus HTTP API, allowing you to query and manage your Prometheus instance through natural language conversations with MCP-compatible clients.

---

## Table of Contents

1. [Features](#features)
2. [Architecture](#architecture)
3. [Requirements](#requirements)
4. [Installation](#installation)
5. [Usage](#usage)
   - [Command Line Flags](#command-line-flags)
6. [API Compatibility](#api-compatibility)
7. [Binding Blocks](#binding-blocks)
   - [Tools](#tools)
   - [Prompts](#prompts)
8. [Contributing](#contributing)
9. [License](#license)
10. [Acknowledgments](#acknowledgments)

---

## Hosted deployment

A hosted deployment is available on [Fronteir AI](https://fronteir.ai/mcp/yshngg-pmcp).

## Features

- **🔥 Golang Implementation**: Built with Go 1.23+ for performance, reliability, and type safety
- **📊 Complete Prometheus API Coverage**: Full compatibility with Prometheus HTTP API v1
- **⚡ Instant Query**: Execute Prometheus queries at a specific point in time
- **📈 Range Query**: Retrieve historical metric data over defined time ranges
- **🔍 Metadata Query**: Discover time series, label names, and label values
- **🎯 Target & Rule Management**: Monitor targets, rules, and alerting configurations
- **🛠️ TSDB Administration**: Advanced database operations including snapshots and series deletion
- **🌐 Multiple Transport Options**: Support for HTTP, Server-Sent Events (SSE), and stdio
- **🤖 MCP Integration**: Seamless communication with MCP-compatible clients like Claude Desktop

---

## Architecture

`prometheus-mcp-server` is designed as a **Golang microservice** that acts as a bridge between MCP clients and Prometheus servers. It provides:

- **Type-safe API bindings** using Go structs that mirror Prometheus API responses
- **Modular package structure** for maintainability and extensibility
- **Comprehensive error handling** with proper Go error propagation
- **Clean separation of concerns** between transport, API client, and business logic

---

## Requirements

- **Go 1.23.0** or higher
- A running **Prometheus server** (v2.x)
- Compatible MCP client (Claude Desktop, custom implementations, etc.)

---

## Installation

### Using Docker (Recommended)

Pull the pre-built image from GitHub Container Registry:

```bash
# Pull the latest image
docker pull ghcr.io/yshngg/prometheus-mcp-server:latest

# Run with stdio transport (for desktop clients)
docker run --rm ghcr.io/yshngg/prometheus-mcp-server:latest --prom-addr="http://host.docker.internal:9090"

# Run with HTTP transport
docker run --rm -p 8080:8080 ghcr.io/yshngg/prometheus-mcp-server:latest --prom-addr="http://host.docker.internal:9090" --transport=http --mcp-addr="0.0.0.0:8080"
```

Alternatively, build locally:

```bash
docker build -t prometheus-mcp-server .
docker run -p 8080:8080 prometheus-mcp-server --prom-addr="http://prometheus:9090" --transport=http
```

### Download Pre-built Binary

Download the latest release from GitHub:

1. Go to `prometheus-mcp-server` [Releases](https://github.com/yshngg/prometheus-mcp-server/releases)
2. Download the appropriate binary for your platform from the **Assets** section
3. Extract and run:

```bash
# Linux/macOS example
tar -xzf prometheus-mcp-server-<version>.linux-amd64.tar.gz
./prometheus-mcp-server --prom-addr="http://localhost:9090"

# Windows example
unzip prometheus-mcp-server-<version>.windows-amd64.zip
prometheus-mcp-server.exe --prom-addr="http://localhost:9090"
```

### Building from Source

```bash
git clone https://github.com/yshngg/prometheus-mcp-server.git
cd prometheus-mcp-server
make build
# Binary will be available as ./prometheus-mcp-server
```

### Using Go Install

Install the `prometheus-mcp-server` binary directly from source:

```bash
go install github.com/yshngg/prometheus-mcp-server@latest
```

Ensure `$GOPATH/bin` is in your `$PATH`.

---

## Usage

Run the server by specifying your Prometheus address and preferred transport:

```bash
# Default (stdio transport) - ideal for desktop clients
prometheus-mcp-server --prom-addr="http://localhost:9090"

# HTTP transport - for web-based integrations
prometheus-mcp-server --prom-addr="http://localhost:9090" --transport=http --mcp-addr="localhost:8080"

# SSE transport - for real-time streaming (deprecated, use HTTP)
prometheus-mcp-server --prom-addr="http://localhost:9090" --transport=sse --mcp-addr="localhost:8080"
```

### Command Line Flags

| Flag         | Description                                       | Default                 |
| ------------ | ------------------------------------------------- | ----------------------- |
| `-help`      | Show help information.                            | N/A                     |
| `-mcp-addr`  | Address for the MCP server to listen on.          | `localhost:8080`        |
| `-prom-addr` | Prometheus server URL.                            | `http://localhost:9090` |
| `-transport` | Communication transport (`stdio`, `http`, `sse`). | `stdio`                 |
| `-version`   | Print version and exit.                           | N/A                     |

---

## API Compatibility

`prometheus-mcp-server` maintains **100% compatibility** with the Prometheus HTTP API v1. Every tool and endpoint corresponds directly to the official Prometheus API:

### Query & Data Retrieval

| Tool          | Prometheus Endpoint   | HTTP Method | Purpose                 |
| ------------- | --------------------- | ----------- | ----------------------- |
| Instant Query | `/api/v1/query`       | GET/POST    | Execute instant queries |
| Range Query   | `/api/v1/query_range` | GET/POST    | Execute range queries   |

### Metadata & Discovery

| Tool                  | Prometheus Endpoint          | HTTP Method | Purpose                          |
| --------------------- | ---------------------------- | ----------- | -------------------------------- |
| Find Series by Labels | `/api/v1/series`             | GET/POST    | Find matching time series        |
| List Label Names      | `/api/v1/labels`             | GET/POST    | List all label names             |
| List Label Values     | `/api/v1/label/:name/values` | GET         | List values for a specific label |
| Target Discovery      | `/api/v1/targets`            | GET         | Get target information           |
| Target Metadata Query | `/api/v1/targets/metadata`   | GET         | Get metadata from targets        |
| Metric Metadata Query | `/api/v1/metadata`           | GET         | Get metric metadata              |

### Rules & Alerts

| Tool                   | Prometheus Endpoint     | HTTP Method | Purpose                      |
| ---------------------- | ----------------------- | ----------- | ---------------------------- |
| Alert Query            | `/api/v1/alerts`        | GET         | Get all active alerts        |
| Rule Query             | `/api/v1/rules`         | GET         | Get recording/alerting rules |
| Alertmanager Discovery | `/api/v1/alertmanagers` | GET         | Get alertmanager information |

### Status & Configuration

| Tool                | Prometheus Endpoint          | HTTP Method | Purpose                   |
| ------------------- | ---------------------------- | ----------- | ------------------------- |
| Config              | `/api/v1/status/config`      | GET         | Get current configuration |
| Flags               | `/api/v1/status/flags`       | GET         | Get runtime flags         |
| Build Information   | `/api/v1/status/buildinfo`   | GET         | Get build information     |
| Runtime Information | `/api/v1/status/runtimeinfo` | GET         | Get runtime information   |
| TSDB Stats          | `/api/v1/status/tsdb`        | GET         | Get TSDB statistics       |
| WAL Replay Stats    | `/api/v1/status/walreplay`   | GET         | Get WAL replay status     |

### TSDB Administration

| Tool             | Prometheus Endpoint                   | HTTP Method | Purpose                 |
| ---------------- | ------------------------------------- | ----------- | ----------------------- |
| TSDB Snapshot    | `/api/v1/admin/tsdb/snapshot`         | POST/PUT    | Create TSDB snapshot    |
| Delete Series    | `/api/v1/admin/tsdb/delete_series`    | POST/PUT    | Delete time series data |
| Clean Tombstones | `/api/v1/admin/tsdb/clean_tombstones` | POST/PUT    | Clean deleted data      |

### Management APIs

| Tool            | Prometheus Endpoint | HTTP Method | Purpose                 |
| --------------- | ------------------- | ----------- | ----------------------- |
| Health Check    | `/-/healthy`        | GET/HEAD    | Check Prometheus health |
| Readiness Check | `/-/ready`          | GET/HEAD    | Check if ready to serve |
| Reload          | `/-/reload`         | PUT/POST    | Reload configuration    |
| Quit            | `/-/quit`           | PUT/POST    | Graceful shutdown       |

**All query parameters, response formats, and error codes match the official Prometheus API specification.**

---

## Binding Blocks

### Tools

**Expression Queries** (Core Prometheus functionality):

- **Instant Query**: Evaluate an instant query at a single point in time
- **Range Query**: Evaluate an expression query over a range of time

**Metadata Queries** (Series and label discovery):

- **Find Series by Labels**: Return the list of time series that match a certain label set
- **List Label Names**: Return a list of label names
- **List Label Values**: Return a list of label values for a provided label name
- **Target Metadata Query**: Return metadata about metrics currently scraped from targets
- **Metric Metadata Query**: Return metadata about metrics currently scraped from targets (without target information)

**Discovery & Monitoring**:

- **Target Discovery**: Return an overview of the current state of the Prometheus target discovery
- **Alert Query**: Return a list of all active alerts
- **Rule Query**: Return a list of alerting and recording rules that are currently loaded
- **Alertmanager Discovery**: Return an overview of the current state of the Prometheus alertmanager discovery

**Status & Configuration**:

- **Config**: Return currently loaded configuration file
- **Flags**: Return flag values that Prometheus was configured with
- **Runtime Information**: Return various runtime information properties about the Prometheus server
- **Build Information**: Return various build information properties about the Prometheus server
- **TSDB Stats**: Return various cardinality statistics about the Prometheus TSDB
- **WAL Replay Stats**: Return information about the WAL replay

**TSDB Admin APIs** (Advanced operations):

- **TSDB Snapshot**: Create a snapshot of all current data into snapshots/`<datetime>`-`<rand>`
- **Delete Series**: Delete data for a selection of series in a time range
- **Clean Tombstones**: Remove the deleted data from disk and cleans up the existing tombstones

**Management APIs**:

- **Health Check**: Check Prometheus health
- **Readiness Check**: Check if Prometheus is ready to serve traffic (i.e. respond to queries)
- **Reload**: Trigger a reload of the Prometheus configuration and rule files
- **Quit**: Trigger a graceful shutdown of Prometheus

### Prompts

- **All Available Metrics**: Return a list of every metric exposed by the Prometheus instance

---

## Contributing

Contributions are welcome! This is a **Golang project**, so please ensure:

- Follow Go best practices and conventions
- Add appropriate tests for new functionality
- Maintain API compatibility with Prometheus
- Update documentation as needed

Please submit a pull request or open an issue to discuss improvements.

### Development Setup

```bash
git clone https://github.com/yshngg/prometheus-mcp-server.git
cd prometheus-mcp-server
go mod download
make build
```

---

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

---


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fyshngg%2Fprometheus-mcp-server.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fyshngg%2Fprometheus-mcp-server?ref=badge_large)

## Acknowledgments

- **Built with Go** using the official [Prometheus Go client library](https://github.com/prometheus/client_golang)
- Powered by [Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Inspired by [Prometheus](https://prometheus.io/) - the de facto standard for monitoring and alerting