package logger

import (
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// MiddlewareOptions can be used to configure the Echo Middleware.
type MiddlewareOptions struct {
	IsIgnorableError func(error) bool
}

const echoKey = "logger"

// Middleware attaches a Logger instance with a request ID onto the context. It
// also logs every request along with metadata about the request.
func Middleware(args ...MiddlewareOptions) func(next echo.HandlerFunc) echo.HandlerFunc {
	var opts MiddlewareOptions
	if len(args) > 0 {
		opts = args[0]
	}

	// by default, we don't ignore any errors
	if opts.IsIgnorableError == nil {
		opts.IsIgnorableError = func(err error) bool {
			return false
		}
	}

	l := New()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t1 := time.Now()

			// create a request ID that will be attached to the
			// logger
			id, err := uuid.NewV4()
			if err != nil {
				return errors.WithStack(err)
			}

			log := l.ID(id.String())
			c.Set(echoKey, log)

			if err := next(c); err != nil {
				if opts.IsIgnorableError(err) {
					log.Err(err).Warn("ignored error")
					return err
				}

				c.Error(err)
			}

			t2 := time.Now()

			// get the last entry in X-Forwarded-For header to
			// determine client IP
			var ipAddress string
			if xff := c.Request().Header.Get("x-forwarded-for"); xff != "" {
				split := strings.Split(xff, ",")
				ipAddress = strings.TrimSpace(split[len(split)-1])
			} else {
				ipAddress = c.Request().RemoteAddr
			}

			log.Root(Data{
				"status_code":   c.Response().Status,
				"method":        c.Request().Method,
				"path":          c.Request().URL.Path,
				"route":         c.Path(),
				"response_time": t2.Sub(t1).Seconds() * 1000,
				"referer":       c.Request().Referer(),
				"user_agent":    c.Request().UserAgent(),
				"ip_address":    ipAddress,
				"trace_id":      c.Request().Header.Get("x-amzn-trace-id"),
			}).Info("handled request")

			return nil
		}
	}
}

// FromEchoContext returns a Logger from the given echo.Context. If there is no
// attached logger, then it will return a new Logger instance.
func FromEchoContext(c echo.Context) Logger {
	if log, ok := c.Get(echoKey).(Logger); ok {
		return log
	}

	return New()
}