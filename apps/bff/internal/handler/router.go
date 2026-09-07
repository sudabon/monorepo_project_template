package handler

import (
	"context"
	"encoding/json/v2"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/identity"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/proxy"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
	"github.com/sudabon/monorepo_project_template/packages/go-platform/echox"
)

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
	e.Use(echox.TraceRequest)
	e.Use(stripClientUserHeader)
	e.Use(loadSession(d.Store))
	e.Use(protectCSRF)
	echox.HealthRoutes(e, d.Ping)
	e.POST("/auth/login", login(d))
	e.POST("/auth/logout", logout(d))
	e.GET("/auth/session", sessionStatus)
	if d.Backend != nil {
		e.Any("/api/*", echo.WrapHandler(proxy.New(d.Backend, identity.UserIDHeader)), requireAuth)
		e.Any("/api", echo.WrapHandler(proxy.New(d.Backend, identity.UserIDHeader)), requireAuth)
	}
	return e
}
