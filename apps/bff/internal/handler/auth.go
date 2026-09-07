package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

const maxLoginBody = 8 << 10

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
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
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
