package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/proxy"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/logging"
	"uuid"
)

const CSRFHeader = "X-CSRF-Token"

const maxLoginBody = 8 << 10

type Deps struct {
	Store        session.Store
	Users        identity.Authenticator
	Backend      *url.URL
	CookieSecure bool
	Ping         func(context.Context) error
}

type jsonSerializer struct{}

func (jsonSerializer) Serialize(c echo.Context, i any, _ string) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	return json.MarshalWrite(c.Response(), i)
}

func (jsonSerializer) Deserialize(c echo.Context, i any) error {
	return json.UnmarshalRead(c.Request().Body, i)
}

func New(d Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.JSONSerializer = jsonSerializer{}
	e.HTTPErrorHandler = handleError
	e.Use(traceRequest)
	e.Use(stripClientUserHeader)
	e.Use(loadSession(d.Store))
	e.Use(protectCSRF)
	e.GET("/health/shallow", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/health/deep", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		status, result := http.StatusOK, "healthy"
		if d.Ping != nil {
			if err := d.Ping(ctx); err != nil {
				status, result = http.StatusServiceUnavailable, "unhealthy"
				slog.WarnContext(ctx, "dependency unhealthy", "dependency", "database", "error", err)
			}
		}
		return c.JSON(status, struct {
			Status       string            `json:"status"`
			Dependencies map[string]string `json:"dependencies"`
		}{result, map[string]string{"database": result}})
	})
	e.POST("/auth/login", login(d))
	e.POST("/auth/logout", logout(d))
	e.GET("/auth/session", sessionStatus)
	if d.Backend != nil {
		e.Any("/api/*", echo.WrapHandler(proxy.New(d.Backend, identity.UserIDHeader)), requireAuth)
		e.Any("/api", echo.WrapHandler(proxy.New(d.Backend, identity.UserIDHeader)), requireAuth)
	}
	return e
}

func traceRequest(next echo.HandlerFunc) echo.HandlerFunc {
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

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type sessionView struct {
	Authenticated bool     `json:"authenticated"`
	User          userView `json:"user,omitzero"`
	CSRFToken     string   `json:"csrfToken,omitempty"`
}

func login(d Deps) echo.HandlerFunc {
	return func(c echo.Context) error {
		// The login endpoint is unauthenticated and internet-facing; cap the
		// body so a large request cannot be read into memory.
		c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxLoginBody)
		var req loginRequest
		if err := c.Bind(&req); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				return echo.NewHTTPError(http.StatusRequestEntityTooLarge)
			}
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		user, err := d.Users.Authenticate(c.Request().Context(), req.Username, req.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		sess, err := d.Store.Create(c.Request().Context(), user.ID, user.Name)
		if err != nil {
			return err
		}
		c.SetCookie(sessionCookie(sess.ID, d.CookieSecure, int(session.IdleTTL.Seconds())))
		return c.JSON(http.StatusOK, sessionView{Authenticated: true, User: userView{ID: user.ID, Name: user.Name}, CSRFToken: sess.CSRFToken})
	}
}

func logout(d Deps) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, ok := session.FromContext(c.Request().Context())
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		if err := d.Store.Delete(c.Request().Context(), sess.ID); err != nil {
			return err
		}
		c.SetCookie(sessionCookie("", d.CookieSecure, -1))
		return c.NoContent(http.StatusNoContent)
	}
}

func sessionStatus(c echo.Context) error {
	sess, ok := session.FromContext(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusOK, sessionView{Authenticated: false})
	}
	return c.JSON(http.StatusOK, sessionView{Authenticated: true, User: userView{ID: sess.UserID, Name: sess.Name}, CSRFToken: sess.CSRFToken})
}

func sessionCookie(value string, secure bool, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     session.CookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		slog.ErrorContext(c.Request().Context(), "response failed after headers were sent", "error", err)
		return
	}
	status, code, message := http.StatusInternalServerError, "internal_error", "An internal error occurred."
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) && httpError.Code >= 400 && httpError.Code < 500 {
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
	if writeErr := c.JSON(status, map[string]string{"code": code, "message": message}); writeErr != nil {
		slog.ErrorContext(c.Request().Context(), "write error response", "error", writeErr)
	}
}
