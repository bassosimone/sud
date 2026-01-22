//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
//

// Package sud provides Single Use Dialers.
//
// A single use dialer allows injecting a pre-established [net.Conn] into
// components that expect to control dialing themselves, such as [http.Transport].
// The first dial succeeds and returns the injected connection; subsequent
// dials fail with [ErrNoConnReuse].
//
// The name "sud" is an acronym for "Single Use Dialer" (and also means
// "south" in Italian, which is a nice coincidence).
package sud

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
)

// NewSingleUseDialer returns a "single use" dialer. The first dial will
// succeed and return the [net.Conn] regardless of the arguments passed to
// the dialing functions. Subsequent dial returns [ErrNoConnReuse].
func NewSingleUseDialer(conn net.Conn) *SingleUseDialer {
	return &SingleUseDialer{conn: conn}
}

// SingleUseDialer is the Dialer returned by [NewSingleUseDialer].
type SingleUseDialer struct {
	mu   sync.Mutex
	conn net.Conn
}

// ErrNoConnReuse is the type of error returned when you create a
// [*SingleUseDialer] and you dial more than once.
var ErrNoConnReuse = errors.New("cannot reuse connection")

// DialContext dials once with the configured connection and then returns [ErrNoConnReuse].
//
// This method signature is compatible with the [net/http] package.
//
// All arguments are ignored and we return the connection (once) or [ErrNoConnRuse].
func (sud *SingleUseDialer) DialContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	sud.mu.Lock()
	defer sud.mu.Unlock()
	if sud.conn == nil {
		return nil, ErrNoConnReuse
	}
	var conn net.Conn
	conn, sud.conn = sud.conn, nil
	return conn, nil
}

// DialTLSContext dials once with the configured connection and then returns [ErrNoConnReuse].
//
// This method signature is compatible with the [golang.org/x/net/http2] package.
//
// All arguments are ignored and we return the connection (once) or [ErrNoConnRuse].
func (d *SingleUseDialer) DialTLSContext(
	ctx context.Context, network, address string, cfg *tls.Config) (net.Conn, error) {
	return d.DialContext(ctx, network, address)
}
