package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	ApiPort              string `mapstructure:"API_PORT"`
	ApiFileUploadMaxSize int    `mapstructure:"API_FILE_UPLOAD_MAX_SIZE"`

	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDBName   string `mapstructure:"POSTGRES_DBNAME"`

	DBMaxOpenConns           int `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBConnMaxLifetimeMinutes int `mapstructure:"DB_CONN_MAX_LIFETIME_MINUTES"`
	DBMaxIdleConns           int `mapstructure:"DB_MAX_IDLE_CONNS"`

	AccessTokenExpiresMinutes int `mapstructure:"ACCESS_TOKEN_EXPIRES_MINUTES"`
	RefreshTokenExpiresHours  int `mapstructure:"REFRESH_TOKEN_EXPIRES_HOURS"`

	JWTSecret string `mapstructure:"JWT_SECRET"`
	JWTIssuer string `mapstructure:"JWT_ISSUER"`

	CookieSecure   bool   `mapstructure:"COOKIE_SECURE"`
	CookieSameSite string `mapstructure:"COOKIE_SAME_SITE"`

	CORSAllowOrigins     string `mapstructure:"CORS_ALLOW_ORIGINS"`
	CORSAllowMethods     string `mapstructure:"CORS_ALLOW_METHODS"`
	CORSAllowHeaders     string `mapstructure:"CORS_ALLOW_HEADERS"`
	CORSAllowCredentials bool   `mapstructure:"CORS_ALLOW_CREDENTIALS"`

	S3Endpoint     string `mapstructure:"MINIO_ENDPOINT"`
	S3AccessKey    string `mapstructure:"MINIO_ACCESS_KEY"`
	S3SecretAccess string `mapstructure:"MINIO_SECRET_KEY"`
	S3Bucket       string `mapstructure:"MINIO_BUCKET"`
	S3UseSSL       bool   `mapstructure:"MINIO_USE_SSL"`
	S3Paginate     int    `mapstructure:"MINIO_FILES_PAGINATE"`

	RedisEndpoint string `mapstructure:"REDIS_ENDPOINT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

func MustNew(envPath string) *Config {
	var config Config

	viper.AutomaticEnv()

	if envPath != "" {
		viper.SetConfigFile(envPath)
		viper.SetConfigType("env")

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("error reading config file from %s: %v", envPath, err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode .env variables into config: %v", err)
	}

	return &config
}
