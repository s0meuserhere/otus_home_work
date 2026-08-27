package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultTimeout = 10 * time.Second

func main() {
	timeout, host, port, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = run(timeout, host, port, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (timeout time.Duration, host string, port string, err error) {
	fs := flag.NewFlagSet("go-telnet", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var fTimeout time.Duration
	fs.DurationVar(&fTimeout, "timeout", defaultTimeout, "connection timeout")

	errParse := fs.Parse(args)
	if errParse != nil {
		return 0, "", "", fmt.Errorf("parse flags: %w", errParse)
	}

	if fs.NArg() != 2 {
		return 0, "", "", fmt.Errorf("usage: go-telnet [--timeout=10s] host port")
	}

	return fTimeout, fs.Arg(0), fs.Arg(1), nil
}

func run(timeout time.Duration, host, port string, in io.ReadCloser, out io.Writer) error {
	address := net.JoinHostPort(host, port)
	client := NewTelnetClient(address, timeout, in, out)

	err := client.Connect()
	if err != nil {
		return err
	}
	defer func() {
		closeErr := client.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "%v\n", closeErr)
		}
	}()

	fmt.Fprintf(os.Stderr, "connected to %s\n", address)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 2)

	go func() {
		err := client.Receive()
		if err == nil {
			fmt.Fprintln(os.Stderr, "connection was closed by peer")
		}

		errCh <- err
	}()

	go func() {
		err := client.Send()
		if err == nil {
			fmt.Fprintln(os.Stderr, "EOF")
		}

		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
