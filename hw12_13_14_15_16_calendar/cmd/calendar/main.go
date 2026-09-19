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
	internalgrpc "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
	"golang.org/x/sync/errgroup"
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

	httpServer := internalhttp.NewServer(logg, cfg.HTTP.Addr(), events)
	grpcServer := internalgrpc.NewServer(logg, cfg.GRPC.Addr(), events)

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()

		if err := httpServer.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server", "err", err)
		}
		if err := grpcServer.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop grpc server", "err", err)
		}
	}()

	logg.Info("calendar is running...", "http", cfg.HTTP.Addr(), "grpc", cfg.GRPC.Addr())

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := httpServer.Start(gCtx); err != nil {
			return fmt.Errorf("http server: %w", err)
		}

		return nil
	})
	g.Go(func() error {
		if err := grpcServer.Start(gCtx); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
