package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/cloud/service/registry"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/web/gateway"
)

// RegisterGateway registers flags for the gateway command.
func RegisterGateway(fs *flag.FlagSet) {
	fs.Int("port", 8080, "listen port for the API gateway")
	fs.String("host", "0.0.0.0", "listen host for the API gateway")
	fs.String("routes", "", "comma-separated route mapping (e.g. /api/auth=auth-service,/api/pay=payment-service)")
}

// RunGateway starts the Krewire API Gateway reverse proxy (KWF-L5H2F FRK-SVC-020).
func RunGateway(fs *flag.FlagSet) errs.ExitCode {
	root, _ := findRoot()
	var cfg *config.Config
	if root != "" {
		if c, err := config.Load(root); err == nil {
			cfg = c
		}
	}

	port := 8080
	if flagProvided(fs, "port") {
		port = flagInt(fs, "port")
	} else if cfg != nil && cfg.Gateway != nil && cfg.Gateway.Port > 0 {
		port = cfg.Gateway.Port
	}

	host := "0.0.0.0"
	if flagProvided(fs, "host") {
		host = flagValue(fs, "host")
	} else if cfg != nil && cfg.Gateway != nil && cfg.Gateway.Host != "" {
		host = cfg.Gateway.Host
	}

	reg := registry.NewMemoryRegistry()
	gw := gateway.New(
		gateway.WithRegistry(reg),
		gateway.WithMiddleware("tracing", gateway.TracingMiddleware()),
		gateway.WithMiddleware("cors", gateway.CORSMiddleware("*")),
	)

	// Register routes from config
	if cfg != nil && cfg.Gateway != nil {
		for _, r := range cfg.Gateway.Routes {
			var timeout time.Duration
			if r.Timeout != "" {
				timeout, _ = time.ParseDuration(r.Timeout)
			}
			gw.AddRoute(gateway.Route{
				Path:        r.Path,
				Method:      r.Method,
				Service:     r.Service,
				TargetURL:   r.TargetURL,
				Prefix:      r.Prefix,
				StripPrefix: r.StripPrefix,
				Middleware:  r.Middleware,
				RateLimit:   r.RateLimit,
				Timeout:     timeout,
			})
		}
	}

	// Register routes from --routes flag (e.g. "/api/v1=service1,/api/v2=service2")
	if flagProvided(fs, "routes") {
		routeSpecs := strings.Split(flagValue(fs, "routes"), ",")
		for _, spec := range routeSpecs {
			parts := strings.SplitN(strings.TrimSpace(spec), "=", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[0])
				target := strings.TrimSpace(parts[1])
				r := gateway.Route{
					Path:   path,
					Prefix: true,
				}
				if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
					r.TargetURL = target
				} else {
					r.Service = target
				}
				gw.AddRoute(r)
			}
		}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:    addr,
		Handler: gw,
	}

	slog.Info("starting API gateway", "addr", addr)
	fmt.Printf("krew gateway listening on http://%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("received signal, shutting down gateway", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		return errs.ExitCodeSuccess
	case err := <-errCh:
		slog.Error("gateway server error", "err", err)
		return fail(err)
	}
}
