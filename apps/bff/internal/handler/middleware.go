package handler

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

const CSRFHeader = "X-CSRF-Token"

func stripClientUserHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().Header.Del(identity.UserIDHeader)
		return next(c)
	}
}

func loadSession(store session.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(session.CookieName)
			if err != nil || cookie.Value == "" {
				return next(c)
			}
			sess, err := store.Get(c.Request().Context(), cookie.Value)
			// Only a missing session means "not signed in". A store outage must
			// surface as 5xx; answering 401 would sign every user out instead.
			if errors.Is(err, session.ErrNotFound) {
				return next(c)
			}
			if err != nil {
				return fmt.Errorf("load session: %w", err)
			}
			c.SetRequest(c.Request().WithContext(session.WithSession(c.Request().Context(), sess)))
			return next(c)
		}
	}
}

func protectCSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !isStateChanging(c.Request().Method) || isLogin(c) {
			return next(c)
		}
		sess, ok := session.FromContext(c.Request().Context())
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		token := c.Request().Header.Get(CSRFHeader)
		if subtle.ConstantTimeCompare([]byte(token), []byte(sess.CSRFToken)) != 1 {
			return echo.NewHTTPError(http.StatusForbidden)
		}
		return next(c)
	}
}

func isStateChanging(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isLogin(c echo.Context) bool {
	return c.Request().Method == http.MethodPost && c.Request().URL.Path == "/auth/login"
}

func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := session.FromContext(c.Request().Context()); !ok {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		return next(c)
	}
}
