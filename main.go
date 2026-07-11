package main

import (
	"context"
	"errors"
	"expvar"
	"flag"
	"fmt"
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
	"k8s.io/klog/v2"
)

const (
	Schema         = "prom"
	methodCallTool = "tools/call"
)

var (
	mcpRequests = expvar.NewInt("mcp_requests_total")
	mcpErrors   = expvar.NewInt("mcp_errors_total")
)

func metricsMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		if method == methodCallTool {
			mcpRequests.Add(1)
		}
		result, err := next(ctx, method, req)
		if method == methodCallTool && err != nil {
			mcpErrors.Add(1)
		}
		return result, err
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	klog.LogToStderr(true)
	defer klog.Flush()

	fs := flag.NewFlagSet("prometheus-mcp-server", flag.ExitOnError)
	klog.InitFlags(fs)
	var (
		promAddr      = fs.String("prom-addr", envOrDefault("PROM_ADDR", "http://localhost:9090/"), "The address of the Prometheus to connect to.")
		mcpAddr       = fs.String("mcp-addr", envOrDefault("MCP_ADDR", "localhost:8080"), "The address of the MCP server to listen on.")
		transportType = fs.String("transport", envOrDefault("TRANSPORT", "stdio"), "Transport type (stdio, sse or http).\nThe mechanisms that handle the underlying communication between clients and servers.")
		printVersion  = fs.Bool("version", false, "Print the version and exit.")
	)
	fs.Usage = usageFor(fs, "prometheus-mcp-server [flags]")
	if err := fs.Parse(os.Args[1:]); err != nil {
		klog.ErrorS(err, "parse args")
		klog.Flush()
		os.Exit(1)
	}

	if *printVersion {
		fmt.Println(version.Info)
		klog.Flush()
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "prometheus-mcp-server",
		Version: string(version.Info.Number),
	}, nil)
	server.AddReceivingMiddleware(metricsMiddleware)

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
		klog.ErrorS(err, "new prometheus client")
		klog.Flush()
		os.Exit(1)
	}

	binder := bindingblocks.NewBinder(server, promCli)
	binder.Bind()

	switch *transportType {
	case "http":
		runHTTP(ctx, server, promCli, *mcpAddr)
	case "sse":
		runSSE(ctx, server, *mcpAddr)
	default:
		runStdio(ctx, server)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func readyzHandler(promCli api.PrometheusAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := promCli.HealthCheck(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not ready","reason":"upstream unhealthy"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

func runHTTP(ctx context.Context, server *mcp.Server, promCli api.PrometheusAPI, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler(promCli))
	mux.Handle("/metrics", expvar.Handler())
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil))

	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			klog.ErrorS(err, "http server shutdown")
		}
	}()

	klog.InfoS("Listening on http", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.ErrorS(err, "listen and serve with Streamable HTTP transport")
		klog.Flush()
		os.Exit(1)
	}
}

func runSSE(ctx context.Context, server *mcp.Server, addr string) {
	klog.InfoS("HTTP+SSE transport is deprecated. Please use Streamable HTTP instead.")

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
			klog.ErrorS(err, "http server shutdown")
		}
	}()

	klog.InfoS("Listening on http", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		klog.ErrorS(err, "listen and serve with HTTP+SSE transport")
		klog.Flush()
		os.Exit(1)
	}
}

func runStdio(ctx context.Context, server *mcp.Server) {
	klog.InfoS("Listening on stdio")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		klog.ErrorS(err, "run server with stdio transport")
		klog.Flush()
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
