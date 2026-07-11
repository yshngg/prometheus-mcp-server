package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	mcpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_requests_total",
			Help: "Total number of MCP tool calls.",
		},
		[]string{"tool"},
	)
	mcpErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_errors_total",
			Help: "Total number of MCP tool call errors.",
		},
		[]string{"tool"},
	)
	mcpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "mcp_request_duration_seconds",
			Help: "Duration of MCP tool calls in seconds.",
		},
		[]string{"tool"},
	)
)

func cacheHintMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		res, err := next(ctx, method, req)
		if err != nil {
			return res, err
		}
		switch r := res.(type) {
		case *mcp.ListToolsResult:
			r.TTLMs = 30000
			r.CacheScope = "public"
		case *mcp.ListPromptsResult:
			r.TTLMs = 30000
			r.CacheScope = "public"
		case *mcp.ListResourcesResult:
			r.TTLMs = 30000
			r.CacheScope = "public"
		case *mcp.ListResourceTemplatesResult:
			r.TTLMs = 30000
			r.CacheScope = "public"
		}
		return res, nil
	}
}

func metricsMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		if method != methodCallTool {
			return next(ctx, method, req)
		}

		toolName := ""
		if req != nil {
			if callReq, ok := req.GetParams().(*mcp.CallToolParams); ok {
				toolName = callReq.Name
			}
		}

		start := time.Now()
		mcpRequests.WithLabelValues(toolName).Inc()
		result, err := next(ctx, method, req)
		mcpDuration.WithLabelValues(toolName).Observe(time.Since(start).Seconds())
		if err != nil {
			mcpErrors.WithLabelValues(toolName).Inc()
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
		transportType = fs.String("transport", envOrDefault("TRANSPORT", "stdio"), "Transport type (stdio or http).\nThe mechanisms that handle the underlying communication between clients and servers.")
		authToken     = fs.String("auth-token", envOrDefault("AUTH_TOKEN", ""), "Bearer token for MCP endpoint authentication. Optional. If empty, no authentication is required.")
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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "prometheus-mcp-server",
		Version: string(version.Info.Number),
	}, &mcp.ServerOptions{
		Instructions: "You are connected to a Prometheus monitoring instance. " +
			"Use expression queries (instant-query, range-query) to explore time-series data. " +
			"Use metadata tools (list-label-names, list-label-values, find-series-by-labels) to discover available metrics and labels. " +
			"Use status tools (config, flags, runtime-information, build-information, tsdb-stats, wal-replay-stats) to inspect server state. " +
			"Resources are available at prom:/// URIs (e.g., prom:///config, prom:///api/v1/query?query=up) for direct data access. " +
			"Use management tools (health-check, readiness-check, reload, quit) with caution as quit and reload affect server operation.",
		SchemaCache: mcp.NewSchemaCache(),
		PageSize:    50,
		CompletionHandler: func(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
			return handleCompletion(ctx, req, promCli)
		},
	})
	server.AddReceivingMiddleware(metricsMiddleware)
	server.AddSendingMiddleware(cacheHintMiddleware)

	binder := bindingblocks.NewBinder(server, promCli)
	binder.Bind()

	switch *transportType {
	case "http":
		runHTTP(ctx, server, promCli, *mcpAddr, *authToken)
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

func authMiddleware(token string) func(http.Handler) http.Handler {
	if token == "" {
		return func(next http.Handler) http.Handler { return next }
	}
	verifier := auth.TokenVerifier(func(_ context.Context, bearer string, _ *http.Request) (*auth.TokenInfo, error) {
		if bearer != token {
			return nil, fmt.Errorf("invalid bearer token: %w", auth.ErrInvalidToken)
		}
		return &auth.TokenInfo{Expiration: time.Now().Add(24 * time.Hour)}, nil
	})
	return auth.RequireBearerToken(verifier, nil)
}

func runHTTP(ctx context.Context, server *mcp.Server, promCli api.PrometheusAPI, addr string, authToken string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler(promCli))
	mux.Handle("/metrics", promhttp.Handler())

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
	mux.Handle("/mcp", authMiddleware(authToken)(mcpHandler))

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

func runStdio(ctx context.Context, server *mcp.Server) {
	klog.InfoS("Listening on stdio")
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		klog.ErrorS(err, "run server with stdio transport")
		klog.Flush()
		os.Exit(1)
	}
}

func handleCompletion(ctx context.Context, req *mcp.CompleteRequest, promCli api.PrometheusAPI) (*mcp.CompleteResult, error) {
	val := req.Params.Argument.Value

	switch req.Params.Ref.Type {
	case "ref/resource":
		uri := req.Params.Ref.URI
		switch {
		case strings.Contains(uri, "label/") && strings.Contains(uri, "/values"):
			names, _, err := promCli.LabelNames(ctx, nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, n := range names {
				if val == "" || strings.HasPrefix(n, val) {
					matches = append(matches, n)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil

		case strings.Contains(uri, "query"):
			values, _, err := promCli.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, v := range values {
				s := string(v)
				if val == "" || strings.HasPrefix(s, val) {
					matches = append(matches, s)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil
		}

	case "ref/prompt":
		if req.Params.Ref.Name == "all-available-metrics" && req.Params.Argument.Name == "prefix" {
			values, _, err := promCli.LabelValues(ctx, "__name__", nil, time.Time{}, time.Time{})
			if err != nil {
				return nil, err
			}
			var matches []string
			for _, v := range values {
				s := string(v)
				if val == "" || strings.HasPrefix(s, val) {
					matches = append(matches, s)
				}
			}
			return &mcp.CompleteResult{
				Completion: mcp.CompletionResultDetails{
					Values:  matches,
					HasMore: len(matches) > 20,
					Total:   len(matches),
				},
			}, nil
		}
	}

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{Values: []string{}},
	}, nil
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
