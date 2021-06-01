package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	t.Run("sets a request ID", func(tt *testing.T) {
		e := echo.New()
		e.Use(Middleware())

		e.GET("/", func(c echo.Context) error {
			log, ok := c.Get("logger").(Logger)
			assert.True(tt, ok)
			assert.NotEmpty(tt, log.id)
			return nil
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(tt, err)

		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		assert.Equal(tt, http.StatusOK, rr.Code)
	})

	t.Run("logs status code", func(tt *testing.T) {
		e := echo.New()
		e.Use(Middleware())

		e.GET("/", func(c echo.Context) error {
			return errors.New("test")
		})

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(tt, err)

		rr := httptest.NewRecorder()

		e.ServeHTTP(rr, req)

		assert.Equal(tt, http.StatusInternalServerError, rr.Code)
	})

	t.Run("does not capture IP address", func(tt *testing.T) {
		e := echo.New()
		out := capturer.CaptureStdout(func() {
			e.Use(Middleware())

			e.GET("/", func(c echo.Context) error {
				return nil
			})

			req, err := http.NewRequest("GET", "/", nil)
			req.Header.Add("x-forwarded-for", "1.1.1.1, 2.2.2.2")
			require.NoError(tt, err)

			rr := httptest.NewRecorder()

			e.ServeHTTP(rr, req)

			assert.Equal(tt, http.StatusOK, rr.Code)
		})

		var data map[string]interface{}
		err := json.Unmarshal([]byte(out), &data)
		require.NoError(tt, err)

		assert.NotContains(tt, data, "ip_address")
	})

	t.Run("ignores errors according to IsIgnorableError", func(tt *testing.T) {
		e := echo.New()
		out := capturer.CaptureStdout(func() {
			e.Use(MiddlewareWithConfig(MiddlewareConfig{
				IsIgnorableError: func(err error) bool {
					return true
				},
			}))

			e.GET("/", func(c echo.Context) error {
				return errors.New("test")
			})

			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(tt, err)

			rr := httptest.NewRecorder()

			e.ServeHTTP(rr, req)

			assert.Equal(tt, http.StatusInternalServerError, rr.Code)
		})

		var data map[string]interface{}
		err := json.Unmarshal([]byte(out), &data)
		require.NoError(tt, err)

		assert.Equal(tt, "ignored error", data["message"])
	})
}

func TestFromEchoContext(t *testing.T) {
	t.Run("retrieves a logger if it has been set previously", func(tt *testing.T) {
		log := New().ID("test")

		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rr := httptest.NewRecorder()
		ctx := e.NewContext(req, rr)

		ctx.Set(echoKey, log)

		l := FromEchoContext(ctx)

		assert.Equal(tt, log.id, l.id)
	})

	t.Run("creates a new logger if one wasn't set previously", func(tt *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rr := httptest.NewRecorder()
		ctx := e.NewContext(req, rr)

		l := FromEchoContext(ctx)

		assert.Empty(tt, l.id)
	})
}
