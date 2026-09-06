package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sudabon/monorepo_project_template/apps/bff/internal/platform/logging"
	"github.com/sudabon/monorepo_project_template/apps/bff/internal/session"
)

func New(backend *url.URL, userHeader string) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(backend)
			pr.Out.Header.Del(userHeader)
			if sess, ok := session.FromContext(pr.In.Context()); ok {
				pr.Out.Header.Set(userHeader, sess.UserID)
			}
			if id := logging.RequestID(pr.In.Context()); id != "" {
				pr.Out.Header.Set(logging.RequestIDHeader, id)
			} else if id := pr.In.Header.Get(logging.RequestIDHeader); id != "" {
				pr.Out.Header.Set(logging.RequestIDHeader, id)
			}
		},
	}
}
