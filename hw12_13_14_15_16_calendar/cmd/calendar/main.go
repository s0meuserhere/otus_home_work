package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/buildinfo"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/infrastructure"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	eventrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/event"
	internalhttp "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.env", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		buildinfo.Print()

		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run calendar: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logg := logger.New(cfg.Logger.Level, cfg.Environment)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var repo eventservice.Repository

	switch cfg.StorageMode {
	case config.ModeMemory:
		repo = eventrepo.NewMemory()
		logg.Info("storage: memory")
	case config.ModeDB:
		dsn, err := cfg.DB.DSN()
		if err != nil {
			return fmt.Errorf("db dsn: %w", err)
		}

		pool, err := infrastructure.NewPgxPool(ctx, dsn)
		if err != nil {
			return fmt.Errorf("db pool: %w", err)
		}
		defer pool.Close()

		repo = eventrepo.NewDB(pool.Pool)
		logg.Info("storage: postgres")
	default:
		return fmt.Errorf("unknown storage_mode: %d", cfg.StorageMode)
	}

	events := eventservice.New(repo)

	server := internalhttp.NewServer(logg, cfg.HTTP.Addr(), events)

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server", "err", err)
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
