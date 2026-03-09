package config

import "github.com/caarlos0/env/v11"

type Config struct {
	ClientID       string `env:"CLIENT_ID,required"`
	ClientSecret   string `env:"CLIENT_SECRET,required"`
	CallbackHost   string `env:"CALLBACK_HOST,required"`
	CallbackURL    string `env:"CALLBACK_URL,required"`
	TokensFile     string `env:"TOKENS_FILE,required"`
	TelegramToken  string `env:"TELEGRAM_TOKEN,required"`
	TelegramChatID int    `env:"TELEGRAM_CHAT_ID,required"`
	ServerHost     string `env:"SERVER_HOST,required"`
}

func New() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
