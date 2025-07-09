package internal

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spacecafe/gobox/gin-authentication"
	"github.com/spacecafe/gobox/gin-ratelimit"
	"github.com/spacecafe/gobox/httpserver"
	"github.com/spacecafe/gobox/logger"
	"github.com/spf13/viper"
)

var (
	ErrNoUpstream     = errors.New("upstream cannot be empty")
	ErrAPIKeysToShort = errors.New("API keys must be at least 16 characters long")
	RegexSplit        = regexp.MustCompile(`[,\s]+`)
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

// NewConfig creates and returns a new Config having default values from the given configuration file.
func NewConfig() *Config {
	config := &Config{
		LogLevel:       "debug",
		HTTPServer:     httpserver.NewConfig(logger.Default()),
		Authentication: authentication.NewConfig(),
		RateLimit:      ratelimit.NewConfig(),
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PROXY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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

	// Load sensitive values from environment variables if available
	if apiKeys, err := loadEnv("AUTHENTICATION_API_KEYS"); err == nil {
		apiKeysList := RegexSplit.Split(apiKeys, -1)
		for i, key := range apiKeysList {
			apiKeysList[i] = strings.TrimSpace(key)
		}
		viper.Set("authentication.api_keys", apiKeysList)
	} else {
		logger.Info(err)
	}

	if users, err := loadEnv("AUTHENTICATION_USERS"); err == nil {
		usersList := RegexSplit.Split(users, -1)
		usersMap := make(map[string]string)
		for _, pair := range usersList {
			parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(parts) == 2 {
				usersMap[parts[0]] = parts[1]
			}
		}
		viper.Set("authentication.users", usersMap)
	} else {
		logger.Info(err)
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

	if r.Upstream == "" {
		return ErrNoUpstream
	}
	if r.upstream, err = url.Parse(r.Upstream); err != nil {
		return err
	}

	if err = r.HTTPServer.Validate(); err != nil {
		return err
	}

	if err = r.Authentication.Validate(); err != nil {
		return err
	}

	// Ensure API keys have sufficient entropy
	for _, key := range r.Authentication.APIKeys {
		if len(key) < 16 {
			return ErrAPIKeysToShort
		}
	}

	if err = r.RateLimit.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *Config) GetUpstream() *url.URL {
	return r.upstream
}

func loadEnv(name string) (string, error) {
	value := viper.GetString(name)
	if value != "" {
		return value, nil
	}

	value = viper.GetString(name + "_FILE")
	if value == "" {
		return "", fmt.Errorf("%s not set", name)
	}
	content, err := os.ReadFile(path.Clean(value))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
