package commands

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/mcp"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/term"
)

// RegisterMCP registers flags for the mcp command.
func RegisterMCP(fs *flag.FlagSet) {
	fs.Bool("stdio", true, "run MCP server over standard input/output (for Cursor, Claude, Antigravity)")
	fs.String("port", "", "run MCP server over HTTP/SSE on the specified port (e.g. 4000)")
	fs.String("host", "127.0.0.1", "host address for HTTP/SSE MCP server")
}

// RunMCP starts the Model Context Protocol (MCP) server for Krewire projects (KWN-MCP-001).
func RunMCP(fs *flag.FlagSet) errs.ExitCode {
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

	port := flagValue(fs, "port")
	if port != "" {
		host := flagValue(fs, "host")
		if host == "" {
			host = "127.0.0.1"
		}
		addr := net.JoinHostPort(host, port)
		mux := http.NewServeMux()
		transport := mcp.NewHTTPTransport(srv)
		transport.Mount(mux, "/mcp")

		httpSrv := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		tm := term.NewTerminal()
		boldCyan := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
		green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
		dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }

		fmt.Printf("\n  %s\n", boldCyan("Krewire MCP Server (HTTP/SSE)"))
		fmt.Printf("  %s http://%s/mcp/sse\n", dim("SSE Endpoint:   "), green(addr))
		fmt.Printf("  %s http://%s/mcp/message\n\n", dim("Message Endpoint:"), addr)
		fmt.Printf("  %s\n\n", dim("Press Ctrl+C to stop."))

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer shutdownCancel()
			_ = httpSrv.Shutdown(shutdownCtx)
		}()

		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fail(err)
		}
		return errs.ExitCodeSuccess
	}

	// Default: Run stdio transport
	if err := mcp.RunStdio(ctx, srv, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}
