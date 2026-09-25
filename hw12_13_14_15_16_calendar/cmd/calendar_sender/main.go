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
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
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
		fmt.Fprintf(os.Stderr, "run calendar_sender: %v\n", err)
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

	logg.Info("calendar_sender stub is running...")

	<-ctx.Done()
	logg.Info("calendar_sender stopped")

	return nil
}
