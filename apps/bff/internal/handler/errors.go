package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/echox"
)

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		slog.ErrorContext(c.Request().Context(), "response failed after headers were sent", "error", err)
		return
	}
	status, code, message := http.StatusInternalServerError, "internal_error", "An internal error occurred."
	if httpError, ok := errors.AsType[*echo.HTTPError](err); ok && httpError.Code >= 400 && httpError.Code < 500 {
		status = httpError.Code
		message = http.StatusText(status)
		switch status {
		case http.StatusBadRequest:
			code = "bad_request"
		case http.StatusUnauthorized:
			code = "unauthenticated"
		case http.StatusForbidden:
			code = "csrf_rejected"
		case http.StatusRequestEntityTooLarge:
			code = "payload_too_large"
		default:
			code = "request_error"
		}
	}
	if status >= 500 {
		slog.ErrorContext(c.Request().Context(), "request failed", "error", err)
	}
	echox.WriteError(c, status, code, message)
}
