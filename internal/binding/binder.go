package binding

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/promapi"
)

type Binder interface {
	Bind()
}

// NewBinder returns a Binder that binds components to the given MCP server using the provided Prometheus client.
func NewBinder(server *mcp.Server, api promapi.PrometheusAPI) Binder {
	return &binder{
		server: server,
		api:    api,
	}
}

type binder struct {
	server *mcp.Server
	api    promapi.PrometheusAPI
}

func (b *binder) Bind() {
	// add tools
	b.addTools()
	// add resources
	b.addResources()
	// add prompts
	b.addPrompts()
}
