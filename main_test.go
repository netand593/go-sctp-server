package main

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

// Mock SCTPConn for testing (implements net.Conn)
type mockSCTPConn struct {
	readData  string
	writeData []byte
	readIndex int
}

func (m *mockSCTPConn) Read(b []byte) (int, error) {
	if m.readIndex >= len(m.readData) {
		return 0, nil // EOF
	}
	n := copy(b, m.readData[m.readIndex:])
	m.readIndex += n
	return n, nil
}

func (m *mockSCTPConn) Write(b []byte) (int, error) {
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockSCTPConn) Close() error { return nil }
func (m *mockSCTPConn) LocalAddr() net.Addr { return &net.TCPAddr{} }
func (m *mockSCTPConn) RemoteAddr() net.Addr { return &net.TCPAddr{} }
func (m *mockSCTPConn) SetDeadline(t time.Time) error { return nil }
func (m *mockSCTPConn) SetReadDeadline(t time.Time) error { return nil }
func (m *mockSCTPConn) SetWriteDeadline(t time.Time) error { return nil }

func TestHandleConnection(t *testing.T) {
	config := &Config{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	mockConn := &mockSCTPConn{readData: "Alice\n"}

	err := handleConnection(context.Background(), mockConn, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "Hello, what's your name? Nice to see you, Alice\n"
	if string(mockConn.writeData) != expected {
		t.Errorf("Expected %q, got %q", expected, string(mockConn.writeData))
	}
}

func TestHandleConnection_InvalidName(t *testing.T) {
	config := &Config{
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	mockConn := &mockSCTPConn{readData: strings.Repeat("a", 101) + "\n"}

	err := handleConnection(context.Background(), mockConn, config)
	if err == nil {
		t.Fatal("Expected error for invalid name length")
	}
}