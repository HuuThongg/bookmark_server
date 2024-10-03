package util

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DATABASE_URL           string        `mapstructure:"DATABASE_URL"`
	MAILGUN_DOMAIN         string        `mapstructure:"mailgunDomain"`
	MailgunAPIKey          string        `mapstructure:"mailgunApiKey"`
	Access_Token_Duration  time.Duration `mapstructure:"accessTokenDuration"`
	Refresh_Token_Duration time.Duration `mapstructure:"refreshTokenDuration"`
	SecretKeyHex           string        `mapstructure:"secretKeyHex"`
	PublicKeyHex           string        `mapstructure:"publicKeyHex"`
	DOSecretKey            string        `mapstructure:"doSecret"`
	DOSpacesKey            string        `mapstructure:"doSpaces"`
	MailJetApiKey          string        `mapstructure:"mailJetApiKey"`
	MailJetSecretKey       string        `mapstructure:"mailJetSecretKey"`
	VultrAccessKey         string        `mapstructure:"vultrAccessKey"`
	VultrSecretKey         string        `mapstructure:"vultrSecretKey"`
	VultrHostname          string        `mapstructure:"vultrHostname"`
	TimeoutRead            time.Duration `mapstructure:"SERVER_TIMEOUT_READ"`
	TimeoutWrite           time.Duration `mapstructure:"SERVER_TIMEOUT_WRITE"`
	TimeoutIdle            time.Duration `mapstructure:"SERVER_TIMEOUT_IDLE"`
	Debug                  bool          `mapstructure:"SERVER_DEBUG"`
	PORT                   string        `mapstructure:"SERVER_PORT"`
	BlackBlazeSecretKey    string        `mapstructure:"blackBlazeSecretKey"`
	BlackBlazeKeyId        string        `mapstructure:"blackBlazeKeyId"`
	BlackBlazeHostName     string        `mapstructure:"BlackBlazeHostName"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path) // <- to work with Dockerfile setup
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.SetConfigFile("config.env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	// Manually check required fields
	requiredFields := []struct {
		name  string
		value string
	}{
		{"DBString", config.DATABASE_URL},
	}

	for _, field := range requiredFields {
		if field.value == "" {
			return config, fmt.Errorf("required field %s is missing", field.name)
		}
	}
	return
}
