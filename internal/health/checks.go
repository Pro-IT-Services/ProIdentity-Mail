package health

import (
	"context"
	"net"
	"time"
)

type CheckResult struct {
	Name string
	OK   bool
	Err  string
}

func TCP(ctx context.Context, name, address string) CheckResult {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return CheckResult{Name: name, OK: false, Err: err.Error()}
	}
	_ = conn.Close()
	return CheckResult{Name: name, OK: true}
}

func Unix(ctx context.Context, name, path string) CheckResult {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "unix", path)
	if err != nil {
		return CheckResult{Name: name, OK: false, Err: err.Error()}
	}
	_ = conn.Close()
	return CheckResult{Name: name, OK: true}
}
