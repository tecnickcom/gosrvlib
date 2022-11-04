package httpclient

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/nexmoinc/gosrvlib/pkg/logging"
	"github.com/nexmoinc/gosrvlib/pkg/redact"
	"github.com/nexmoinc/gosrvlib/pkg/traceid"
	"github.com/nexmoinc/gosrvlib/pkg/uidc"
	"go.uber.org/zap"
)

// Client wraps the default HTTP client functionalities and adds logging and instrumentation capabilities.
type Client struct {
	client            *http.Client
	component         string
	logPrefix         string
	traceIDHeaderName string
	redactFn          RedactFn
}

// defaultClient() returns a default client.
func defaultClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout:   1 * time.Minute,
			Transport: http.DefaultTransport,
		},
		traceIDHeaderName: traceid.DefaultHeader,
		component:         "-",
		redactFn:          redact.HTTPData,
	}
}

// New creates a new HTTP client instance.
func New(opts ...Option) *Client {
	c := defaultClient()

	for _, applyOpt := range opts {
		applyOpt(c)
	}

	return c
}

// Do performs the HTTP request with added trace ID, logging and metrics.
//
//nolint:gocognit
func (c *Client) Do(r *http.Request) (resp *http.Response, err error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(r.Context(), c.client.Timeout)
	defer cancel()

	l := logging.FromContext(ctx).With(zap.String(c.logPrefix+"component", c.component))
	debug := l.Check(zap.DebugLevel, "debug") != nil

	defer func() {
		l = l.With(zap.Duration(c.logPrefix+"duration", time.Since(start)))

		if err != nil {
			l.Error(c.logPrefix+"error", zap.Error(err))
			return
		}

		if debug {
			l.Debug(c.logPrefix + "outbound")
			return
		}

		l.Info(c.logPrefix + "outbound")
	}()

	reqID := traceid.FromContext(ctx, uidc.NewID128())
	ctx = traceid.NewContext(ctx, reqID)
	r.Header.Set(c.traceIDHeaderName, reqID)
	r = r.WithContext(ctx)

	l = l.With(
		zap.String(c.logPrefix+"traceid", reqID),
		zap.String(c.logPrefix+"request_method", r.Method),
		zap.String(c.logPrefix+"request_path", r.URL.Path),
		zap.String(c.logPrefix+"request_query", r.URL.RawQuery),
		zap.String(c.logPrefix+"request_uri", r.RequestURI),
	)

	if debug {
		reqDump, errd := httputil.DumpRequestOut(r, true)
		if errd == nil {
			l = l.With(
				zap.String(c.logPrefix+"request", c.redactFn(string(reqDump))),
			)
		}
	}

	resp, err = c.client.Do(r)

	if debug && resp != nil {
		respDump, errd := httputil.DumpResponse(resp, true)
		if errd == nil {
			l = l.With(
				zap.String(c.logPrefix+"response", c.redactFn(string(respDump))),
			)
		}
	}

	return resp, err //nolint:wrapcheck
}
