// Package echox holds Echo-specific wiring shared by the API and the BFF.
// Importing this package is what pulls Echo into a binary; logging, server,
// and database stay independent.
package echox

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
	"uuid"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
)

func TraceRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		id := c.Request().Header.Get(logging.RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		ctx := logging.WithRequestID(c.Request().Context(), id)
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		req := c.Request().Clone(ctx)
		req.Header.Set(logging.RequestIDHeader, id)
		c.SetRequest(req)
		c.Response().Header().Set(logging.RequestIDHeader, id)
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				// Keep the panic site in the error so the 500 log names it.
				// handleError never puts the error text in the response body.
				err = fmt.Errorf("request panic: %v\n%s", recovered, debug.Stack())
			}
			if err != nil {
				c.Error(err)
				err = nil
			}
			slog.InfoContext(ctx, "request completed", "method", c.Request().Method, "path", c.Path(), "status", c.Response().Status, "duration_ms", time.Since(start).Milliseconds())
		}()
		return next(c)
	}
}

func HealthRoutes(e *echo.Echo, ping func(context.Context) error) {
	// ALB must use shallow: a DB outage must not recycle every healthy task.
	e.GET("/health/shallow", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/health/deep", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		status, result := http.StatusOK, "healthy"
		if ping != nil {
			if err := ping(ctx); err != nil {
				status, result = http.StatusServiceUnavailable, "unhealthy"
				slog.WarnContext(ctx, "dependency unhealthy", "dependency", "database", "error", err)
			}
		}
		return c.JSON(status, struct {
			Status       string            `json:"status"`
			Dependencies map[string]string `json:"dependencies"`
		}{result, map[string]string{"database": result}})
	})
}

func WriteError(c echo.Context, status int, code, message string) {
	if c.Response().Committed {
		slog.ErrorContext(c.Request().Context(), "response failed after headers were sent")
		return
	}
	if writeErr := c.JSON(status, map[string]string{"code": code, "message": message}); writeErr != nil {
		slog.ErrorContext(c.Request().Context(), "write error response", "error", writeErr)
	}
}
