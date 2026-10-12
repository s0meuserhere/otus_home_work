package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/buildinfo"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/infrastructure"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	notifyrepo "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/repository/notify"
	senderservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/sender"
	"golang.org/x/sync/errgroup"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/sender_config.env", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		buildinfo.Print()

		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run calendar_sender: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadSender(configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	queue := notifyrepo.NewRabbit(client)
	sender := senderservice.New(notifyrepo.NewDB(pool.Pool))

	logg.Info("calendar_sender is running...", "queue", cfg.Rabbit.Queue)

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return queue.Consume(gCtx, cfg.Sender.Prefetch, sender.Send)
	})
	g.Go(func() error {
		return client.Wait(gCtx)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	logg.Info("calendar_sender stopped")

	return nil
}
