package commands

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/mcp"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/tui/term"
)

// RegisterMCP registers flags for the mcp command.
func RegisterMCP(fs *flag.FlagSet) {
	fs.Bool("stdio", false, "run MCP server over standard input/output (default mode when --port is not specified)")
	fs.String("port", "", "run MCP server over HTTP/SSE on the specified port (e.g. 4000)")
	fs.String("host", "127.0.0.1", "host address for HTTP/SSE MCP server (loopback default)")
	fs.String("auth-token", "", "authentication token required for HTTP/SSE MCP requests (mandatory for non-loopback binds)")
	fs.String("token", "", "alias for --auth-token")
	fs.String("allowed-origins", "", "comma-separated list of additional allowed HTTP Origin values")
	fs.String("allowed-hosts", "", "comma-separated list of additional allowed Host header values")
}

// isLoopback checks if a given host address resolves to a local loopback interface.
func isLoopback(host string) bool {
	h := strings.TrimSpace(host)
	if parsedHost, _, err := net.SplitHostPort(h); err == nil {
		h = parsedHost
	}
	h = strings.Trim(strings.ToLower(h), "[]")
	if h == "127.0.0.1" || h == "localhost" || h == "::1" || h == "" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// RunMCP starts the Model Context Protocol (MCP) server for Krewire projects (KWN-MCP-001).
func RunMCP(fs *flag.FlagSet) errs.ExitCode {
	port := flagValue(fs, "port")
	stdioProvided := flagProvided(fs, "stdio")
	portProvided := flagProvided(fs, "port") && port != ""

	if stdioProvided && portProvided {
		return fail(fmt.Errorf("cannot specify both --stdio and --port: choose either standard I/O or HTTP transport"))
	}

	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	cfg, _ := config.Load(root)
	srv := mcp.NewServer(root, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	if portProvided {
		host := flagValue(fs, "host")
		if host == "" {
			host = "127.0.0.1"
		}

		authToken := flagValue(fs, "auth-token")
		if authToken == "" {
			authToken = flagValue(fs, "token")
		}
		if authToken == "" {
			authToken = os.Getenv("KREW_MCP_TOKEN")
		}

		loopback := isLoopback(host)
		if !loopback && authToken == "" {
			return fail(fmt.Errorf("refusing to bind MCP server to non-loopback address %q without --auth-token (or KREW_MCP_TOKEN)", host))
		}

		var allowedOrigins []string
		if origins := flagValue(fs, "allowed-origins"); origins != "" {
			for _, o := range strings.Split(origins, ",") {
				if s := strings.TrimSpace(o); s != "" {
					allowedOrigins = append(allowedOrigins, s)
				}
			}
		}

		var allowedHosts []string
		if host != "" {
			allowedHosts = append(allowedHosts, host)
		}
		if extraHosts := flagValue(fs, "allowed-hosts"); extraHosts != "" {
			for _, h := range strings.Split(extraHosts, ",") {
				if s := strings.TrimSpace(h); s != "" {
					allowedHosts = append(allowedHosts, s)
				}
			}
		}

		secOpts := mcp.SecurityOptions{
			AllowedOrigins: allowedOrigins,
			AllowedHosts:   allowedHosts,
			AuthToken:      authToken,
		}

		addr := net.JoinHostPort(host, port)
		mux := http.NewServeMux()
		transport := mcp.NewHTTPTransport(srv, secOpts)
		transport.Mount(mux, "/mcp")

		httpSrv := &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			IdleTimeout:       120 * time.Second,
		}

		tm := term.NewTerminal()
		boldCyan := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
		green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
		dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }
		yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, []term.Style{term.StyleBold}) }

		fmt.Printf("\n  %s\n", boldCyan("Krewire MCP Server (HTTP/SSE)"))
		fmt.Printf("  %s http://%s/mcp/sse\n", dim("SSE Endpoint:     "), green(addr))
		fmt.Printf("  %s http://%s/mcp/message\n", dim("Message Endpoint: "), addr)

		if !loopback {
			fmt.Fprintf(os.Stderr, "\n  %s %s\n", yellow("[WARNING]"), tm.Paint(fmt.Sprintf("Bound to non-loopback address %s. Tool execution is network-accessible.", addr), term.ColorYellow, nil))
			fmt.Printf("  %s %s\n", dim("Auth Enforcement: "), green("Bearer Token Required"))
		} else if authToken != "" {
			fmt.Printf("  %s %s\n", dim("Auth Enforcement: "), green("Bearer Token Required"))
		} else {
			fmt.Printf("  %s %s\n", dim("Security Mode:    "), dim("Loopback only (Origin & Host validation enabled)"))
		}

		fmt.Printf("\n  %s\n\n", dim("Press Ctrl+C to stop."))

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			_ = httpSrv.Shutdown(shutdownCtx)
		}()

		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fail(err)
		}
		return errs.ExitCodeSuccess
	}

	// Default or explicit STDIO transport
	if err := mcp.RunStdio(ctx, srv, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}
