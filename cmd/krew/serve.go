package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/transport"
)

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.Int("port", DefaultServePort, "Port to listen on")
	hostFlag := fs.String("host", DefaultServeHost, "Host address to listen on")
	personaFlag := fs.String("persona", DefaultPersonaName, "Default agent persona")
	workspaceFlag := fs.String("workspace", DefaultWorkspaceDir, "Workspace directory")
	noAuthFlag := fs.Bool("no-auth", false, "Disable authentication for local loopback development")
	_ = fs.Parse(args)

	pers := parsePersona(*personaFlag)

	inst, err := embed.New(embed.Options{
		Workspace: *workspaceFlag,
		Persona:   &pers,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing agent: %v\n", err)
		os.Exit(1)
	}

	server := transport.NewHTTPServer(inst.Agent(), transport.WithNoAuth(*noAuthFlag))
	addr := fmt.Sprintf("%s:%d", *hostFlag, *portFlag)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Handler(),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("\033[1;36m[Krew Server]\033[0m Listening on http://%s (persona: %s)\n", addr, pers.Name)
		fmt.Println("Available endpoints:")
		fmt.Printf("  GET  http://%s/health\n", addr)
		fmt.Printf("  POST http://%s/api/agent/run  (stream=true for SSE)\n", addr)
		fmt.Printf("  POST http://%s/api/agent/chat\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	fmt.Println("Server stopped.")
}
