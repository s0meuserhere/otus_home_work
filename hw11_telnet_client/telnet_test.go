package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connection refused", func(t *testing.T) {
		client := NewTelnetClient("127.0.0.1:1", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		err := client.Connect()
		require.Error(t, err)
	})

	t.Run("connection timeout", func(t *testing.T) {
		client := NewTelnetClient("254.254.254.254:80", 50*time.Millisecond, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
		err := client.Connect()
		require.Error(t, err)
	})

	t.Run("peer closes connection", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		go func() {
			conn, acceptErr := l.Accept()
			require.NoError(t, acceptErr)
			require.NoError(t, conn.Close())
		}()

		out := &bytes.Buffer{}
		client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(&bytes.Buffer{}), out)
		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		err = client.Receive()
		require.NoError(t, err)
		require.Empty(t, out.String())
	})

	t.Run("send until eof", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		received := make(chan string, 1)

		go func() {
			conn, acceptErr := l.Accept()
			require.NoError(t, acceptErr)
			defer func() { require.NoError(t, conn.Close()) }()

			data, readErr := io.ReadAll(conn)
			require.NoError(t, readErr)
			received <- string(data)
		}()

		in := bytes.NewBufferString("hello from client")
		client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(in), &bytes.Buffer{})
		require.NoError(t, client.Connect())

		err = client.Send()
		require.NoError(t, err)
		require.NoError(t, client.Close())

		select {
		case got := <-received:
			require.Equal(t, "hello from client", got)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for server to receive data")
		}
	})

	t.Run("not connected", func(t *testing.T) {
		client := NewTelnetClient("127.0.0.1:0", time.Second, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})

		err := client.Send()
		require.Error(t, err)

		err = client.Receive()
		require.Error(t, err)

		require.NoError(t, client.Close())
	})
}
