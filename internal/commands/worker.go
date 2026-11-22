package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/cloud/storage"
	"github.com/krewire/krewire/pkg/cloud/worker"
	"github.com/krewire/krewire/pkg/errs"
)

// RegisterWorker registers flags for the worker command.
func RegisterWorker(fs *flag.FlagSet) {
	fs.String("queue", "memory", "queue backend: memory, file, or kv (default memory)")
	fs.Int("concurrency", 0, "number of concurrent workers (default: config or GOMAXPROCS)")
	fs.String("dlq", "", "inspect or manage DLQ: list, clear")
}

// RunWorker starts background workers for the current project (KWF-L5H2F FRK-SVC-060/061).
func RunWorker(fs *flag.FlagSet) errs.ExitCode {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}

	dlqOp := flagValue(fs, "dlq")
	if dlqOp == "" && cfg.Worker == nil && !hasFile(root, "main.go") {
		fmt.Fprintln(os.Stderr, "krew worker: requires 'worker:' configuration in krewire.yaml or main.go")
		return errs.ExitCodeUsage
	}

	queueBackend := flagValue(fs, "queue")
	if queueBackend == "" {
		queueBackend = "memory"
	}

	var q worker.Queue
	switch queueBackend {
	case "file", "kv":
		storeDir := filepath.Join(root, ".krewire", "storage")
		fileKV, err := storage.NewFile(storeDir)
		if err != nil {
			return fail(err)
		}
		kvq, err := worker.NewKVQueue(ctx, fileKV)
		if err != nil {
			return fail(err)
		}
		q = kvq
	default:
		q = worker.NewInMemoryQueue()
	}

	dlqCmd := flagValue(fs, "dlq")
	if dlqCmd != "" {
		return runDLQ(ctx, q, dlqCmd)
	}

	concurrency := flagInt(fs, "concurrency")
	if concurrency <= 0 && cfg.Worker != nil && cfg.Worker.Concurrency > 0 {
		concurrency = cfg.Worker.Concurrency
	}

	var opts []worker.RunnerOption
	if concurrency > 0 {
		opts = append(opts, worker.WithConcurrency(concurrency))
	}
	runner := worker.NewRunner(q, opts...)

	slog.Info("worker runner started", "backend", queueBackend, "concurrency", concurrency)

	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		slog.Info("worker shutting down")
		runner.Stop()
		<-errCh
		return errs.ExitCodeSuccess
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("worker runner error", "error", err)
			return errs.ExitCodeFailure
		}
		return errs.ExitCodeSuccess
	}
}

// runDLQ handles DLQ inspection and management commands (KWF-L5H2F FRK-SVC-062).
func runDLQ(ctx context.Context, q worker.Queue, cmd string) errs.ExitCode {
	switch cmd {
	case "list":
		dlq := q.DLQ()
		if len(dlq) == 0 {
			fmt.Println("DLQ is empty (0 dead letters)")
			return errs.ExitCodeSuccess
		}
		fmt.Printf("DLQ contains %d dead letter(s):\n", len(dlq))
		for i, dl := range dlq {
			fmt.Printf(" [%d] ID: %s | Attempts: %d | Error: %v | At: %s\n", i+1, dl.ID, dl.Attempts, dl.Err, dl.At.Format("2006-01-02 15:04:05"))
		}
		return errs.ExitCodeSuccess
	case "clear":
		if err := q.ClearDLQ(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "krew worker dlq: failed to clear DLQ: %v\n", err)
			return errs.ExitCodeFailure
		}
		fmt.Println("DLQ cleared successfully")
		return errs.ExitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "krew worker dlq: unknown sub-command %q (supported: list, clear)\n", cmd)
		return errs.ExitCodeUsage
	}
}
