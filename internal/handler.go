package internal

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func handleProxy(upstreamURL *url.URL, authHeader string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Remove potentially sensitive headers
			req.Header.Del("Authorization")
			req.Header.Del(authHeader)

			// Add proxy information
			if req.Header.Get("X-Forwarded-Host") == "" {
				req.Header.Set("X-Forwarded-Host", req.Host)
			}
			if req.Header.Get("X-Forwarded-Proto") == "" {
				req.Header.Set("X-Forwarded-Proto", req.Proto)
			}
			if clientIP, _, err := net.SplitHostPort(ctx.Request.RemoteAddr); err == nil {
				if forwardedIP := req.Header.Get("X-Forwarded-For"); forwardedIP == "" {
					req.Header.Set("X-Forwarded-For", clientIP)
				} else {
					req.Header.Set("X-Forwarded-For", forwardedIP+", "+clientIP)
				}
			}
		}

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
