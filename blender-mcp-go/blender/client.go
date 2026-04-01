package blender

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const Addr = "127.0.0.1:9876"

type Client struct {
	addr string
}

func New() *Client {
	return &Client{addr: Addr}
}

func (c *Client) Send(cmd string, params map[string]any) (map[string]any, error) {
	return c.SendWithTimeout(3*time.Second, cmd, params)
}

func (c *Client) SendWithTimeout(timeout time.Duration, cmd string, params map[string]any) (map[string]any, error) {
	conn, err := net.DialTimeout("tcp", c.addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("bridge not running on %s — click Start Bridge in Blender N-panel", c.addr)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	data, _ := json.Marshal(map[string]any{"cmd": cmd, "params": params})
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	buf := make([]byte, 1<<17)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	var res map[string]any
	if err := json.Unmarshal(buf[:n], &res); err != nil {
		return nil, fmt.Errorf("bad JSON from Blender: %w", err)
	}
	if e, ok := res["error"].(string); ok && e != "" {
		return nil, fmt.Errorf("%s", e)
	}
	return res, nil
}
