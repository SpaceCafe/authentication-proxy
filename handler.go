package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func handleProxy(upstreamURL *url.URL) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
		proxy.Director = func(req *http.Request) {
			req.Header = ctx.Request.Header
			req.Host = upstreamURL.Host
			req.URL.Scheme = upstreamURL.Scheme
			req.URL.Host = upstreamURL.Host
			req.URL.Path = ctx.Param("proxyPath")
		}
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
