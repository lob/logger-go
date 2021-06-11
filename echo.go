package logger

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// MiddlewareConfig can be used to configure the Echo Middleware.
type MiddlewareConfig struct {
	IsIgnorableError func(error) bool
}

var defaultMiddlewareConfig = MiddlewareConfig{
	IsIgnorableError: func(err error) bool {
		return false
	},
}

const echoKey = "logger"

// Middleware attaches a Logger instance with a request ID onto the context. It
// also logs every request along with metadata about the request. To customize
// the middleware, use MiddlewareWithConfig.
func Middleware(serviceName string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return MiddlewareWithConfig(serviceName, defaultMiddlewareConfig)
}

// MiddlewareWithConfig attaches a Logger instance with a request ID onto the
// context. It also logs every request along with metadata about the request.
// Pass in a MiddlewareConfig to customize the behavior of the middleware.
func MiddlewareWithConfig(serviceName string, opts MiddlewareConfig) func(next echo.HandlerFunc) echo.HandlerFunc {
	l := New(serviceName)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t1 := time.Now()

			// create a request ID that will be attached to the
			// logger
			id, err := uuid.NewV4()
			if err != nil {
				return errors.WithStack(err)
			}

			log := l.ID(id.String()).Root(Data{
				"method":     c.Request().Method,
				"route":      c.Path(),
				"path":       c.Request().URL.Path,
				"trace_id":   c.Request().Header.Get("x-amzn-trace-id"),
				"referer":    c.Request().Referer(),
				"user_agent": c.Request().UserAgent(),
			})
			c.Set(echoKey, log)

			if err := next(c); err != nil {
				if opts.IsIgnorableError(err) {
					log.Err(err).Warn("ignored error")
					return err
				}

				c.Error(err)
			}

			t2 := time.Now()

			log.Root(Data{
				"status_code":   c.Response().Status,
				"response_time": t2.Sub(t1).Seconds() * 1000,
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

	return New("")
}
