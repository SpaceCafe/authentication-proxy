package main

import (
	authentication "github.com/spacecafe/gobox/gin-authentication"
	ratelimit "github.com/spacecafe/gobox/gin-ratelimit"
	"github.com/spacecafe/gobox/httpserver"
	"github.com/spacecafe/gobox/logger"
	"github.com/spacecafe/gobox/terminator"
)

func main() {
	term := terminator.New(terminator.NewConfig())
	config := NewConfig()

	server := httpserver.NewHTTPServer(config.HTTPServer)
	server.Engine.Use(ratelimit.New(config.RateLimit))
	server.Engine.Use(authentication.New(config.Authentication))
	server.Router.Any("/*proxyPath", handleProxy(config.GetUpstream()))
	if err := server.Start(term.FullTracking()); err != nil {
		logger.Fatal(err)
	}

	term.Wait()
}
