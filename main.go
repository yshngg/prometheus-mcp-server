package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yshngg/prometheus-mcp-server/internal/bindingblocks"
	"github.com/yshngg/prometheus-mcp-server/internal/prometheus/api"
	"github.com/yshngg/prometheus-mcp-server/internal/version"
)

const Schema = "prom"

func main() {
	fs := flag.NewFlagSet("prometheus-mcp-server", flag.ExitOnError)
	var (
		promAddr      = fs.String("prom-addr", "http://localhost:9090/", "The address of the Prometheus to connect to.")
		mcpAddr       = fs.String("mcp-addr", "localhost:8080", "The address of the MCP server to listen on.")
		transportType = fs.String("transport", "stdio", "Transport type (stdio, sse or http).\nThe mechanisms that handle the underlying communication between clients and servers.")
		printVersion  = fs.Bool("version", false, "Print the version and exit.")
	)
	fs.Usage = usageFor(fs, "prometheus-mcp-server [flags]")
	if err := fs.Parse(os.Args[1:]); err != nil {
		slog.Error("parse args", "err", err)
		os.Exit(1)
	}

	if *printVersion {
		fmt.Println(version.Info)
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "prometheus-mcp-server",
		Version: string(version.Info.Number),
	}, nil)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	promCli, err := api.New(*promAddr, httpClient, nil)
	if err != nil {
		slog.Error("new prometheus client", "err", err)
		os.Exit(1)
	}

	binder := bindingblocks.NewBinder(server, promCli)
	binder.Bind()

	switch *transportType {
	case "http":
		runHTTP(ctx, server, *mcpAddr)
	case "sse":
		runSSE(ctx, server, *mcpAddr)
	default:
		runStdio(ctx, server)
	}
}

func runHTTP(ctx context.Context, server *mcp.Server, addr string) {
	mux := http.NewServeMux()
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil))

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown", "err", err)
		}
	}()

	slog.Info("Listening on http://" + addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("listen and serve with Streamable HTTP transport", "err", err)
		os.Exit(1)
	}
}

func runSSE(ctx context.Context, server *mcp.Server, addr string) {
	slog.Warn("HTTP+SSE transport is deprecated. Please use Streamable HTTP instead.")

	mux := http.NewServeMux()
		mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	mux.Handle("/mcp", mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return server
	}, nil))

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown", "err", err)
		}
	}()

	slog.Info("Listening on http://" + addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("listen and serve with HTTP+SSE transport", "err", err)
		os.Exit(1)
	}
}

func runStdio(ctx context.Context, server *mcp.Server) {
	slog.Info("Listening on stdio")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		slog.Error("run server with stdio transport", "err", err)
		os.Exit(1)
	}
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Prometheus Model Context Protocol Server\n\n")
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			def := f.DefValue
			if def == "" {
				def = "..."
			}
			_, err := fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, def, f.Usage)
			if err != nil {
				panic(err)
			}
		})
		if err := w.Flush(); err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "VERSION\n")
		fmt.Fprintf(os.Stderr, "  %s\n", version.Info.Number)
		fmt.Fprintf(os.Stderr, "\n")
	}
}
