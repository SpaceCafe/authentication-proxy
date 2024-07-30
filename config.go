package main

import (
	"errors"
	"flag"
	"net/url"

	authentication "github.com/spacecafe/gobox/gin-authentication"
	ratelimit "github.com/spacecafe/gobox/gin-ratelimit"
	"github.com/spacecafe/gobox/httpserver"
	"github.com/spacecafe/gobox/logger"
	"github.com/spf13/viper"
)

var (
	ErrNoUpstream = errors.New("upstream cannot be empty")
)

// Config defines the essential parameters for serving this application.
type Config struct {
	LogLevel       string `json:"log_level" yaml:"log_level" mapstructure:"log_level"`
	Upstream       string `json:"upstream" yaml:"upstream" mapstructure:"upstream"`
	upstream       *url.URL
	HTTPServer     *httpserver.Config     `json:"http_server" yaml:"http_server" mapstructure:"http_server"`
	Authentication *authentication.Config `json:"authentication" yaml:"authentication" mapstructure:"authentication"`
	RateLimit      *ratelimit.Config      `json:"rate_limit" yaml:"rate_limit" mapstructure:"rate_limit"`
}

// NewConfig creates and returns a new Config having default values from given configuration file.
func NewConfig() *Config {
	config := &Config{
		LogLevel:       "debug",
		HTTPServer:     httpserver.NewConfig(logger.Default()),
		Authentication: authentication.NewConfig(),
		RateLimit:      ratelimit.NewConfig(),
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	configPath := flag.String("config", "", "Path to config.yaml")
	flag.Parse()
	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	} else {
		viper.AddConfigPath("/etc/authentication-proxy/")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal(err)
	}

	err = viper.Unmarshal(config)
	if err != nil {
		logger.Fatal(err)
	}

	if err = config.Validate(); err != nil {
		logger.Fatal(err)
	}

	return config
}

// Validate ensures the all necessary configurations are filled and within valid confines.
// Any misconfiguration results in well-defined standardized errors.
func (r *Config) Validate() error {
	var err error
	if err = logger.ParseLevel(r.LogLevel); err != nil {
		return err
	}
	if r.upstream, err = url.Parse(r.Upstream); r.Upstream == "" || err != nil {
		return ErrNoUpstream
	}
	if err = r.HTTPServer.Validate(); err != nil {
		return err
	}
	if err = r.Authentication.Validate(); err != nil {
		return err
	}
	if err = r.RateLimit.Validate(); err != nil {
		return err
	}
	return nil
}

func (r *Config) GetUpstream() *url.URL {
	return r.upstream
}
