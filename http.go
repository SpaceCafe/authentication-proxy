package main

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var upstreamURL *url.URL

// StartServer starts a web server on the given IP address and port
func StartServer() error {
	var err error

	if upstreamURL, err = url.Parse(fmt.Sprintf("http://%s:%d", viper.GetString("upstream.host"), viper.GetInt("upstream.port"))); err != nil {
		panic(err)
	}

	// Initialize route
	http.HandleFunc("/", HandlerRateLimit(HandlerAuthentication(httputil.NewSingleHostReverseProxy(upstreamURL).ServeHTTP), viper.GetInt("server.max_connections")))

	// Initialize http server
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
		ReadHeaderTimeout: time.Minute,
	}

	return server.ListenAndServe()
}

// HandlerRateLimit is HTTP handling middleware that queues and rate limits client requests to ensures
// no more than `max_connections` requests are passed concurrently to the given handler.
func HandlerRateLimit(h http.HandlerFunc, max int) http.HandlerFunc {
	sema := make(chan struct{}, max)

	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case sema <- struct{}{}:
			defer func() { <-sema }()
			h(w, r)
		case <-time.After(viper.GetDuration("server.timeout")):
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}
	}
}

// HandlerAuthentication search and check API authentication headers
func HandlerAuthentication(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Check for header
		if r.Header.Get("X-API-Key") == viper.GetString("server.api_key") || r.Header.Get("API-Key") == viper.GetString("server.api_key") {
			return
		}

		if _, password, ok := r.BasicAuth(); ok && password == viper.GetString("server.api_key") {
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorised"))
	}
}
