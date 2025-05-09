package ami

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
)

// Socket holds the socket client connection data.
type Socket struct {
	conn     net.Conn
	incoming chan string
	shutdown chan struct{}
	errors   chan error
}

// NewSocket provides a new socket client, connecting to a tcp server.
func NewSocket(ctx context.Context, address string) (*Socket, error) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	s := &Socket{
		conn:     conn,
		incoming: make(chan string, 32),
		shutdown: make(chan struct{}),
		errors:   make(chan error, 2),
	}
	go s.run(ctx, conn)
	return s, nil
}

// Connected returns the socket status, true for connected,
// false for disconnected.
func (s *Socket) Connected() bool {
	return s.conn != nil
}

// Close closes socket connection.
func (s *Socket) Close(ctx context.Context) error {
	close(s.shutdown)
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// Send sends data to socket using fprintf format.
func (s *Socket) Send(message string) error {
	_, err := fmt.Fprintf(s.conn, message)
	return err
}

// Recv receives a string from socket server.
func (s *Socket) Recv(ctx context.Context) (string, error) {
	var buffer bytes.Buffer
	for {
		select {
		case msg, ok := <-s.incoming:
			if !ok {
				return buffer.String(), io.EOF
			}
			buffer.WriteString(msg)
			if strings.HasSuffix(buffer.String(), "\r\n") {
				return buffer.String(), nil
			}
		case err := <-s.errors:
			return buffer.String(), err
		case <-s.shutdown:
			return buffer.String(), io.EOF
		case <-ctx.Done():
			return buffer.String(), io.EOF
		}
	}
}

func (s *Socket) run(ctx context.Context, conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			s.errors <- err
			return
		}
		s.incoming <- msg
	}
}
