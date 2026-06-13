// package broadcast provides a thin client for the Liquidsoap control interface.
//
// ✮ ⋆ ˚｡𖦹 how it works:
//
// liquidsoap exposes a unix socket at /var/run/liquidsoap/radio.sock (configured
// in radio.liq). both the liquidsoap container and the api container mount the same
// host directory (code/streaming/socket/) so they share the socket file without
// any network exposure.
//
// the protocol is simple line-based text. each command is a single line; the
// response is one or more lines terminated by a bare "END" line:
//
//	→  radio_queue.push /media/track.mp3\n
//	←  4\n
//	←  END\n
//
// the response "4" is the request ID liquidsoap assigned to the queued file.
// liquidsoap processes one command at a time, so Client serialises calls via a mutex.
//
// the only command the broadcast controller needs right now is Push, which
// queues a file onto the named request.queue source defined in radio.liq.
// as the controller grows, new commands can be added via Command directly.
package broadcast

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Client is a thin TCP client for the Liquidsoap telnet interface.
// Each command is serialised via mu — Liquidsoap processes one command at a time.
type Client struct {
	addr   string
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

// Dial connects to the Liquidsoap unix socket at path (e.g. "/var/run/liquidsoap/radio.sock").
func Dial(path string) (*Client, error) {
	conn, err := net.DialTimeout("unix", path, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("liquidsoap dial: %w", err)
	}
	return &Client{
		addr:   path,
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// Command sends cmd and reads the response up to the "END" terminator.
func (c *Client) Command(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ⋆˙⟡ liquidsoap responses end with a bare "END" line
	if err := c.conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return "", fmt.Errorf("liquidsoap deadline: %w", err)
	}

	if _, err := fmt.Fprintf(c.conn, "%s\n", cmd); err != nil {
		return "", fmt.Errorf("liquidsoap write: %w", err)
	}

	var sb strings.Builder
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("liquidsoap read: %w", err)
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "END" {
			break
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(trimmed)
	}

	return sb.String(), nil
}

// Push queues path onto queueID (the id= set in radio.liq).
// Returns the request ID Liquidsoap assigned, e.g. "4".
func (c *Client) Push(queueID, path string) (string, error) {
	resp, err := c.Command(fmt.Sprintf("%s.push %s", queueID, path))
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(resp, "ERROR") {
		return "", fmt.Errorf("liquidsoap: %s", resp)
	}
	return resp, nil
}

// Close sends quit and closes the connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, _ = fmt.Fprintf(c.conn, "quit\n")
	return c.conn.Close()
}
