package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{address: address, timeout: timeout, in: in, out: out}
}

// Connect устанавливает TCP-соединение с удалённым адресом.
func (c *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.address, err)
	}

	c.conn = conn

	return nil
}

// Close закрывает TCP-соединение.
func (c *telnetClient) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}

// Send копирует данные из входного потока в сокет.
func (c *telnetClient) Send() error {
	err := c.copy(c.conn, c.in, "send")
	if err != nil {
		return err
	}

	tcpConn, ok := c.conn.(*net.TCPConn)
	if !ok {
		return nil
	}

	err = tcpConn.CloseWrite()
	if err != nil {
		return fmt.Errorf("close write: %w", err)
	}

	return nil
}

// Receive копирует данные из сокета в выходной поток.
func (c *telnetClient) Receive() error {
	err := c.copy(c.out, c.conn, "receive")
	if err != nil {
		return err
	}

	return nil
}

func (c *telnetClient) copy(dst io.Writer, src io.Reader, op string) error {
	if c.conn == nil {
		return fmt.Errorf("%s: not connected", op)
	}

	_, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
