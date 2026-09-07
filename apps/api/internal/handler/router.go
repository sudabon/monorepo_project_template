package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/generated"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/echox"
)

func New(items *usecase.Items, ping func(context.Context) error) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = handleError
	e.Use(echox.TraceRequest)
	generated.RegisterHandlersWithBaseURL(e, &Items{usecase: items}, "/api")
	echox.HealthRoutes(e, ping)
	return e
}

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		slog.ErrorContext(c.Request().Context(), "response failed after headers were sent", "error", err)
		return
	}
	status, code, message := http.StatusInternalServerError, "internal_error", "An internal error occurred."
	if fields, ok := errors.AsType[domain.ValidationErrors](err); ok {
		fieldErrors := make([]generated.FieldError, 0, len(fields))
		for _, f := range fields {
			fieldErrors = append(fieldErrors, generated.FieldError{Field: f.Field, Message: f.Message})
		}
		if writeErr := c.JSON(http.StatusUnprocessableEntity, generated.ValidationError{Code: "validation_error", Message: "Some fields are invalid.", Errors: fieldErrors}); writeErr != nil {
			slog.ErrorContext(c.Request().Context(), "write validation response", "error", writeErr)
		}
		return
	}
	httpError, isHTTP := errors.AsType[*echo.HTTPError](err)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "Item not found."
	case isHTTP && httpError.Code >= 400 && httpError.Code < 500:
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
	echox.WriteError(c, status, code, message)
}
