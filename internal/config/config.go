package config

import "github.com/caarlos0/env/v11"

type Config struct {
	BaseURL                 string `env:"BASE_URL,required"`
	IsProduction            bool   `env:"IS_PRODUCTION,required"`
	JWTSecret               string `env:"JWT_SECRET,required"`
	GithubClientID          string `env:"GITHUB_CLIENT_ID,required"`
	GithubClientSecret      string `env:"GITHUB_CLIENT_SECRET,required"`
	DatabaseURL             string `env:"DATABASE_URL,required"`
	UserSyncInterval        int    `env:"USER_SYNC_INTERVAL,required"`
	LoginSuccessRedirectURL string `env:"LOGIN_SUCCESS_REDIRECT_URL,required"`
}

func ParseConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
