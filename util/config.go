package util

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DATABASE_URL                 string        `mapstructure:"DATABASE_URL"`
	Access_Token_Duration        time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	Refresh_Token_Duration       time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	SecretKeyHex                 string        `mapstructure:"SECRET_KEY_HEX"`
	PublicKeyHex                 string        `mapstructure:"PUBLIC_KEY_HEX"`
	DOSecretKey                  string        `mapstructure:"DO_SECRET"`
	DOSpacesKey                  string        `mapstructure:"DO_SPACES"`
	MailJetApiKey                string        `mapstructure:"MAILJET_API_KEY"`
	MailJetSecretKey             string        `mapstructure:"MAILJET_SECRET_KEY"`
	TimeoutRead                  time.Duration `mapstructure:"SERVER_TIMEOUT_READ"`
	TimeoutWrite                 time.Duration `mapstructure:"SERVER_TIMEOUT_WRITE"`
	TimeoutIdle                  time.Duration `mapstructure:"SERVER_TIMEOUT_IDLE"`
	Debug                        bool          `mapstructure:"SERVER_DEBUG"`
	PORT                         string        `mapstructure:"SERVER_PORT"`
	BlackBlazeSecretKey          string        `mapstructure:"BLACKBLAZE_SECRET_KEY"`
	BlackBlazeKeyId              string        `mapstructure:"BLACKBLAZE_KEY_ID"`
	BlackBlazeHostName           string        `mapstructure:"BLACKBLAZE_HOSTNAME"`
	HOST                         string        `mapstructure:"HOST"`
	DOMAIN                       string        `mapstructure:"DOMAIN"`
	VALKEY_URL                   string        `mapstructure:"VALKEY_URL"`
	CLOUDFLARE_ACCESS_KEY_ID     string        `mapstructure:"CLOUDFLARE_ACCESS_KEY_ID"`
	CLOUDFLARE_ENDPOINT          string        `mapstructure:"CLOUDFLARE_ENDPOINT"`
	CLOUDFLARE_SECRET_ACCESS_KEY string        `mapstructure:"CLOUDFLARE_SECRET_ACCESS_KEY"`
	CLOUDFLARE_OG_BUCKET         string        `mapstructure:"CLOUDFLARE_OG_BUCKET"`
	CLOUDFLARE_REGION            string        `mapstructure:"CLOUDFLARE_REGION"`
}

func LoadConfig(path string) (config Config, err error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found or error loading .env file")
	}

	config.DATABASE_URL = os.Getenv("DATABASE_URL")
	// config.MAILGUN_DOMAIN = os.Getenv("MAILGUN_DOMAIN")
	// config.MailgunAPIKey = os.Getenv("MAILGUN_API_KEY")
	accessTokenDuration := os.Getenv("ACCESS_TOKEN_DURATION")
	refreshTokenDuration := os.Getenv("REFRESH_TOKEN_DURATION")
	config.SecretKeyHex = os.Getenv("SECRET_KEY_HEX")
	config.PublicKeyHex = os.Getenv("PUBLIC_KEY_HEX")
	config.DOSecretKey = os.Getenv("DO_SECRETS")
	config.DOSpacesKey = os.Getenv("DO_SPACES")
	config.MailJetApiKey = os.Getenv("MAILJET_API_KEY")
	config.MailJetSecretKey = os.Getenv("MAILJET_SECRET_KEY")
	timeoutRead := os.Getenv("SERVER_TIMEOUT_READ")
	timeoutWrite := os.Getenv("SERVER_TIMEOUT_WRITE")
	timeoutIdle := os.Getenv("SERVER_TIMEOUT_IDLE")
	debug := os.Getenv("SERVER_DEBUG")
	config.PORT = os.Getenv("SERVER_PORT")
	config.BlackBlazeSecretKey = os.Getenv("BLACKBLAZE_SECRET_KEY")
	config.BlackBlazeKeyId = os.Getenv("BLACKBLAZE_KEY_ID")
	config.BlackBlazeHostName = os.Getenv("BLACKBLAZE_HOSTNAME")
	config.DOMAIN = os.Getenv("DOMAIN")
	config.CLOUDFLARE_ACCESS_KEY_ID = os.Getenv("CLOUDFLARE_ACCESS_KEY_ID")
	config.CLOUDFLARE_ENDPOINT = os.Getenv("CLOUDFLARE_ENDPOINT")
	config.CLOUDFLARE_SECRET_ACCESS_KEY = os.Getenv("CLOUDFLARE_SECRET_ACCESS_KEY")
	config.CLOUDFLARE_OG_BUCKET = os.Getenv("CLOUDFLARE_OG_BUCKET")
	config.CLOUDFLARE_REGION = os.Getenv("CLOUDFLARE_REGION")
	config.VALKEY_URL = os.Getenv("VALKEY_URL")
	if config.VALKEY_URL == "" {
		log.Panic("VALKEY_URL is empty")
	}
	if config.DOMAIN == "" {
		config.DOMAIN = "http://localhost:8080"
	}
	config.HOST = os.Getenv("HOST")
	if config.HOST == "" {
		config.HOST = "http://localhost:5173"
	}

	// Convert string durations to time.Duration
	if accessTokenDuration != "" {
		if duration, err := time.ParseDuration(accessTokenDuration); err == nil {
			config.Access_Token_Duration = duration
		}
	}

	if refreshTokenDuration != "" {
		if duration, err := time.ParseDuration(refreshTokenDuration); err == nil {
			config.Refresh_Token_Duration = duration
		}
	}

	if timeoutRead != "" {
		if duration, err := time.ParseDuration(timeoutRead); err == nil {
			config.TimeoutRead = duration
		}
	}

	if timeoutWrite != "" {
		if duration, err := time.ParseDuration(timeoutWrite); err == nil {
			config.TimeoutWrite = duration
		}
	}

	if timeoutIdle != "" {
		if duration, err := time.ParseDuration(timeoutIdle); err == nil {
			config.TimeoutIdle = duration
		}
	}

	// Convert debug string to boolean
	if debug == "true" {
		config.Debug = true
	} else {
		config.Debug = false
	}

	// Check required fields using reflection
	val := reflect.ValueOf(config)
	typ := reflect.TypeOf(config)

	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i).Interface()
		fieldName := typ.Field(i).Tag.Get("mapstructure") // Get the field name from the struct tag

		// Check if the field is a string and empty
		if strVal, ok := fieldValue.(string); ok && strVal == "" {
			return config, fmt.Errorf("required field %s is missing", fieldName)
		}

		// For time.Duration, you may want to check if it is 0 or another default
		if durationVal, ok := fieldValue.(time.Duration); ok && durationVal == 0 {
			return config, fmt.Errorf("required field %s is missing", fieldName)
		}
	}

	return config, nil
}
