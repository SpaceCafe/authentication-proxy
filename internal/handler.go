package internal

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func handleProxy(upstreamURL *url.URL) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
