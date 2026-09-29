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
	notifyrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/notify"
	schedulerservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/scheduler"
	"golang.org/x/sync/errgroup"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/scheduler_config.env", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		buildinfo.Print()

		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run calendar_scheduler: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadScheduler(configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.StorageMode != config.ModeDB {
		return fmt.Errorf("storage_mode must be %d (db), got %d", config.ModeDB, cfg.StorageMode)
	}

	logg := logger.New(cfg.Logger.Level, cfg.Environment)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	ctx = logger.WithContext(ctx, logg.Slog())

	dsn, err := cfg.DB.DSN()
	if err != nil {
		return fmt.Errorf("db dsn: %w", err)
	}

	pool, err := infrastructure.NewPgxPool(ctx, dsn)
	if err != nil {
		return fmt.Errorf("db pool: %w", err)
	}
	defer pool.Close()

	client, err := infrastructure.NewRabbitClient(cfg.Rabbit)
	if err != nil {
		return fmt.Errorf("rabbit: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			logg.Error("failed to close rabbit client", "err", err)
		}
	}()

	scheduler := schedulerservice.New(
		eventrepo.NewDB(pool.Pool),
		notifyrepo.NewDB(pool.Pool),
		notifyrepo.NewRabbit(client),
		cfg.Scheduler.Retention,
	)

	logg.Info("calendar_scheduler is running...",
		"interval", cfg.Scheduler.Interval,
		"retention", cfg.Scheduler.Retention,
		"run_timeout", cfg.Scheduler.RunTimeout)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return runPeriodically(gCtx, scheduler, cfg.Scheduler.Interval, cfg.Scheduler.RunTimeout)
	})
	g.Go(func() error {
		return client.Wait(gCtx)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	logg.Info("calendar_scheduler stopped")

	return nil
}

// runPeriodically запускает планировщик сразу и затем каждые interval до отмены ctx.
func runPeriodically(
	ctx context.Context,
	scheduler schedulerservice.Service,
	interval time.Duration,
	runTimeout time.Duration,
) error {
	if interval <= 0 {
		return fmt.Errorf("invalid scheduler interval: %s", interval)
	}
	if runTimeout <= 0 {
		return fmt.Errorf("invalid scheduler run timeout: %s", runTimeout)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Без этой проверки после сигнала мог начаться ещё один запуск.
		if ctx.Err() != nil {
			return nil
		}

		runOnce(ctx, scheduler, runTimeout)

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// runOnce доводит запуск до конца даже после сигнала, чтобы уведомления не ушли повторно.
func runOnce(ctx context.Context, scheduler schedulerservice.Service, timeout time.Duration) {
	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	if err := scheduler.SendPending(runCtx, now); err != nil {
		log.Error("scheduler: send pending", "err", err)
	}

	if err := scheduler.DeleteOld(runCtx, now); err != nil {
		log.Error("scheduler: delete old", "err", err)
	}
}
