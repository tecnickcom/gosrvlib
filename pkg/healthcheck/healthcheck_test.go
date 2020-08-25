package healthcheck

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testHealthChecker struct {
	delay time.Duration
	err   error
}

func (th *testHealthChecker) HealthCheck(ctx context.Context) error {
	if th.delay != 0 {
		time.Sleep(th.delay)
	}
	return th.err
}

func TestNew(t *testing.T) {
	hc := &testHealthChecker{}
	h := New("hc-id_1", hc)

	require.NotNil(t, h)
	require.Equal(t, h.ID, "hc-id_1")
	require.Equal(t, h.Checker, hc)
	require.Equal(t, h.Timeout, DefaultTimeout)
}

func TestNewWithTimeout(t *testing.T) {
	hc := &testHealthChecker{}
	to := 5 * time.Second
	h := NewWithTimeout("hc-id_1", hc, to)

	require.NotNil(t, h)
	require.Equal(t, h.ID, "hc-id_1")
	require.Equal(t, h.Checker, hc)
	require.Equal(t, h.Timeout, to)
}
