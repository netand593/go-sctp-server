package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ishidawataru/sctp"
	"golang.org/x/sync/semaphore"
)

// Config holds server configuration
type Config struct {
	IP           string
	Port         int
	MaxConn      int64
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// handleConnection manages a single SCTP connection with improved error handling and validation
func handleConnection(ctx context.Context, conn net.Conn, config *Config) error {
	defer conn.Close()

	// Set timeouts
	conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))

	reader := bufio.NewReader(conn)

	// Send greeting
	message := "Hello, what's your name? "
	if _, err := conn.Write([]byte(message)); err != nil {
		slog.Error("Failed to send greeting", "error", err)
		return err
	}

	// Read name
	name, err := reader.ReadString('\n')
	if err != nil {
		slog.Error("Failed to read from client", "error", err)
		return err
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 100 {
		err := fmt.Errorf("invalid name length: %d", len(name))
		slog.Error("Invalid input", "error", err)
		return err
	}

	// Send response
	response := fmt.Sprintf("Nice to see you, %s\n", name)
	if _, err := conn.Write([]byte(response)); err != nil {
		slog.Error("Failed to send response", "error", err)
		return err
	}

	slog.Info("Handled connection successfully", "name", name)
	return nil
}

func main() {
	config := &Config{}

	// Command-line flags
	flag.StringVar(&config.IP, "ip", "0.0.0.0", "IP address to bind to")
	flag.IntVar(&config.Port, "port", 38412, "Port to listen on")
	flag.Int64Var(&config.MaxConn, "max-conn", 100, "Maximum concurrent connections")
	flag.DurationVar(&config.ReadTimeout, "read-timeout", 30*time.Second, "Read timeout")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", 30*time.Second, "Write timeout")
	flag.Parse()

	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	IPAddr := sctp.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.ParseIP(config.IP)}}, Port: config.Port}

	listener, err := sctp.ListenSCTP("sctp", &IPAddr)
	if err != nil {
		slog.Error("Failed to create SCTP listener", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("Listening for SCTP connections", "address", IPAddr.String())

	// Semaphore for limiting concurrent connections
	sem := semaphore.NewWeighted(config.MaxConn)

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down server")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		sctpConn := conn.(*sctp.SCTPConn)

		// Acquire semaphore
		if err := sem.Acquire(context.Background(), 1); err != nil {
			slog.Warn("Too many connections, dropping", "error", err)
			conn.Close()
			continue
		}

		go func() {
			defer sem.Release(1)
			ctx := context.Background()
			if err := handleConnection(ctx, sctpConn, config); err != nil {
				slog.Error("Connection handling failed", "error", err)
			}
		}()
	}
}
