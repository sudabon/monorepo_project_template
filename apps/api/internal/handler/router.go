package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/generated"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/platform/logging"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
)

func New(items *usecase.Items, ping func(context.Context) error) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = handleError
	e.Use(traceRequest)
	generated.RegisterHandlersWithBaseURL(e, &Items{usecase: items}, "/api")
	// ALB must use shallow: a DB outage must not recycle every healthy task.
	e.GET("/health/shallow", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "healthy"}) })
	e.GET("/health/deep", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		status, result := http.StatusOK, "healthy"
		if err := ping(ctx); err != nil {
			status, result = http.StatusServiceUnavailable, "unhealthy"
			slog.WarnContext(ctx, "dependency unhealthy", "dependency", "database", "error", err)
		}
		return c.JSON(status, struct {
			Status       string            `json:"status"`
			Dependencies map[string]string `json:"dependencies"`
		}{result, map[string]string{"database": result}})
	})
	return e
}

func traceRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		id := c.Request().Header.Get(logging.RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		ctx := logging.WithRequestID(c.Request().Context(), id)
		// Bound DB/query work without tying in-flight requests to the signal context.
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		c.SetRequest(c.Request().WithContext(ctx))
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

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		slog.ErrorContext(c.Request().Context(), "response failed after headers were sent", "error", err)
		return
	}
	status, code, message := http.StatusInternalServerError, "internal_error", "An internal error occurred."
	var fields domain.ValidationErrors
	var httpError *echo.HTTPError
	if errors.As(err, &fields) {
		errors := make([]generated.FieldError, 0, len(fields))
		for _, f := range fields {
			errors = append(errors, generated.FieldError{Field: f.Field, Message: f.Message})
		}
		if writeErr := c.JSON(http.StatusUnprocessableEntity, generated.ValidationError{Code: "validation_error", Message: "Some fields are invalid.", Errors: errors}); writeErr != nil {
			slog.ErrorContext(c.Request().Context(), "write validation response", "error", writeErr)
		}
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "Item not found."
	case errors.As(err, &httpError) && httpError.Code >= 400 && httpError.Code < 500:
		status = httpError.Code
		message = http.StatusText(status)
		switch status {
		case 400:
			code = "bad_request"
		case 404:
			code = "not_found"
		case 405:
			code = "method_not_allowed"
		case 415:
			code = "unsupported_media_type"
		default:
			code = "request_error"
		}
	}
	if status >= 500 {
		slog.ErrorContext(c.Request().Context(), "request failed", "error", err)
	}
	if writeErr := c.JSON(status, generated.Error{Code: code, Message: message}); writeErr != nil {
		slog.ErrorContext(c.Request().Context(), "write error response", "error", writeErr)
	}
}
