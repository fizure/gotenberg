package context

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thecodingmachine/gotenberg/internal/app/api/pkg/resource"
	"github.com/thecodingmachine/gotenberg/internal/pkg/config"
	"github.com/thecodingmachine/gotenberg/internal/pkg/logger"
	"github.com/thecodingmachine/gotenberg/internal/pkg/standarderror"
)

// Context extends the default echo.Context.
type Context struct {
	echo.Context
	logger    *logger.Logger
	config    *config.Config
	resource  *resource.Resource
	startTime time.Time
}

// New creates a new context.
func New(c echo.Context, logger *logger.Logger, config *config.Config) *Context {
	// TODO timeout context?
	return &Context{
		c,
		logger,
		config,
		nil,
		time.Now(),
	}
}

// MustCastFromEchoContext cast an echo.Context to our custom
// context. If something goes wrong, panic.
func MustCastFromEchoContext(c echo.Context) *Context {
	ctx, ok := c.(*Context)
	if !ok {
		panic("unable to cast an echo.Context to a custom context")
	}
	return ctx
}

// StandardLogger returns the custom logger.
// This method should be used instead of the
// default Logger() method coming from
// the echo.Context!
func (ctx *Context) StandardLogger() *logger.Logger {
	return ctx.logger
}

// Resource returns the associated resource
// to the context.
func (ctx *Context) Resource() *resource.Resource {
	return ctx.resource
}

// WithResource adds a resource to the context.
func (ctx *Context) WithResource(resourceDirPath string) error {
	const op = "context.WithResource"
	r, err := resource.New(ctx, ctx.logger, ctx.config, resourceDirPath)
	ctx.resource = r
	if err != nil {
		return &standarderror.Error{
			Op:  op,
			Err: err,
		}
	}
	return nil
}

// LogRequestResult logs the result of a request.
// This method should only be used by a middleware!
func (ctx *Context) LogRequestResult(err error, isDebug bool) error {
	req := ctx.Request()
	resp := ctx.Response()
	stopTime := time.Now()
	fields := map[string]interface{}{
		"time_rfc3339":  timeRFC3339(), // FIXME required?
		"remote_ip":     ctx.RealIP(),
		"host":          req.Host,
		"uri":           req.RequestURI,
		"method":        req.Method,
		"path":          path(req),
		"referer":       req.Referer(),
		"user_agent":    req.UserAgent(),
		"status":        resp.Status,
		"latency":       lantency(ctx.startTime, stopTime),
		"latency_human": latencyHuman(ctx.startTime, stopTime),
		"bytes_in":      bytesIn(req),
		"bytes_out":     bytesOut(resp),
	}
	if err != nil {
		ctx.logger.WithFields(fields).Error("request failed")
		return err
	}
	if isDebug {
		ctx.logger.WithFields(fields).Debug("request handled")
		return nil
	}
	ctx.logger.WithFields(fields).Info("request handled")
	return nil
}

func timeRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

func path(r *http.Request) string {
	path := r.URL.Path
	if path == "" {
		path = "/"
	}
	return path
}

func lantency(startTime time.Time, stopTime time.Time) string {
	return strconv.FormatInt(int64(stopTime.Sub(startTime)), 10)
}

func latencyHuman(startTime time.Time, stopTime time.Time) string {
	return stopTime.Sub(startTime).String()
}

func bytesIn(r *http.Request) string {
	bytesIn := r.Header.Get(echo.HeaderContentLength)
	if bytesIn == "" {
		bytesIn = "0"
	}
	return bytesIn
}

func bytesOut(r *echo.Response) string {
	return strconv.FormatInt(r.Size, 10)
}
