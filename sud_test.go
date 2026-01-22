// SPDX-License-Identifier: GPL-3.0-or-later

package sud

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSingleUseDialer(t *testing.T) {
	conn := &net.TCPConn{}
	dialer := NewSingleUseDialer(conn)
	require.NotNil(t, dialer)
	assert.Equal(t, conn, dialer.conn)
}

func TestSingleUseDialer_DialContext(t *testing.T) {
	t.Run("first dial succeeds", func(t *testing.T) {
		// Arrange: create a dialer with a mock connection
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)

		// Act: dial once
		got, err := dialer.DialContext(context.Background(), "tcp", "example.com:443")

		// Assert: should succeed and return the injected connection
		require.NoError(t, err)
		assert.Equal(t, conn, got)
	})

	t.Run("second dial fails with ErrNoConnReuse", func(t *testing.T) {
		// Arrange: create a dialer and consume the connection
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)
		_, err := dialer.DialContext(context.Background(), "tcp", "example.com:443")
		require.NoError(t, err)

		// Act: dial again
		got, err := dialer.DialContext(context.Background(), "tcp", "example.com:443")

		// Assert: should fail with ErrNoConnReuse
		assert.Nil(t, got)
		assert.ErrorIs(t, err, ErrNoConnReuse)
	})

	t.Run("arguments are ignored", func(t *testing.T) {
		// Arrange: create a dialer
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)

		// Act: dial with different arguments than what the connection was for
		got, err := dialer.DialContext(context.Background(), "udp", "different.host:8080")

		// Assert: should still return the injected connection
		require.NoError(t, err)
		assert.Equal(t, conn, got)
	})
}

func TestSingleUseDialer_DialTLSContext(t *testing.T) {
	t.Run("delegates to DialContext", func(t *testing.T) {
		// Arrange: create a dialer with a mock connection
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)

		// Act: dial with DialTLSContext
		got, err := dialer.DialTLSContext(context.Background(), "tcp", "example.com:443", &tls.Config{})

		// Assert: should return the injected connection
		require.NoError(t, err)
		assert.Equal(t, conn, got)
	})

	t.Run("second dial fails with ErrNoConnReuse", func(t *testing.T) {
		// Arrange: create a dialer and consume the connection via DialTLSContext
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)
		_, err := dialer.DialTLSContext(context.Background(), "tcp", "example.com:443", nil)
		require.NoError(t, err)

		// Act: dial again
		got, err := dialer.DialTLSContext(context.Background(), "tcp", "example.com:443", nil)

		// Assert: should fail with ErrNoConnReuse
		assert.Nil(t, got)
		assert.ErrorIs(t, err, ErrNoConnReuse)
	})
}

func TestErrNoConnReuse(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		assert.Equal(t, "cannot reuse connection", ErrNoConnReuse.Error())
	})

	t.Run("errors.Is detection", func(t *testing.T) {
		// Arrange: get the error from a second dial
		conn := &net.TCPConn{}
		dialer := NewSingleUseDialer(conn)
		_, _ = dialer.DialContext(context.Background(), "tcp", "example.com:443")

		// Act
		_, err := dialer.DialContext(context.Background(), "tcp", "example.com:443")

		// Assert
		assert.True(t, errors.Is(err, ErrNoConnReuse))
	})
}

func TestSingleUseDialer_ConcurrentAccess(t *testing.T) {
	// Arrange: create a dialer with a mock connection
	conn := &net.TCPConn{}
	dialer := NewSingleUseDialer(conn)

	// Act: attempt to dial concurrently from multiple goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	successes := make(chan net.Conn, numGoroutines)
	failures := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			got, err := dialer.DialContext(context.Background(), "tcp", "example.com:443")
			if err != nil {
				failures <- err
			} else {
				successes <- got
			}
		}()
	}
	wg.Wait()
	close(successes)
	close(failures)

	// Assert: exactly one goroutine should succeed
	var successCount int
	for c := range successes {
		assert.Equal(t, conn, c)
		successCount++
	}
	assert.Equal(t, 1, successCount)

	// Assert: all other goroutines should fail with ErrNoConnReuse
	var failureCount int
	for err := range failures {
		assert.ErrorIs(t, err, ErrNoConnReuse)
		failureCount++
	}
	assert.Equal(t, numGoroutines-1, failureCount)
}
