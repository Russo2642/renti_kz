package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	DefaultTimezone = "Asia/Almaty"

	DatabaseTimezone = "UTC"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	S3           S3Config
	JWT          JWTConfig
	App          AppConfig
	Migration    MigrationConfig
	Tuya         TuyaConfig
	Redis        RedisConfig
	Notification NotificationConfig
	OTP          OTPConfig
	FreedomPay   FreedomPayConfig
	Log          LogConfig
}

type ServerConfig struct {
	HttpPort          string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int
	EnableKeepAlive   bool
	KeepAlivePeriod   time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

type S3Config struct {
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type AppConfig struct {
	PasswordSalt string
	Environment  string
	Timezone     string
}

type TuyaConfig struct {
	ClientID     string
	ClientSecret string
	APIBase      string
	TimeZone     string
}

func (c *AppConfig) GetLocation() (*time.Location, error) {
	if c.Timezone == "" {
		c.Timezone = DefaultTimezone
	}
	return time.LoadLocation(c.Timezone)
}

type MigrationConfig struct {
	AutoMigrate bool

	MigrationsPath string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int

	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type NotificationConfig struct {
	RedisQueueName string
	PollInterval   time.Duration
}

type OTPConfig struct {
	Token    string
	APIBase  string
	From     string
	Template string
}

type FreedomPayConfig struct {
	MerchantID string
	SecretKey  string
	APIBase    string
	WebhookURL string
}

type LogConfig struct {
	Level      string `json:"level"`       // "debug", "info", "warn", "error"
	Format     string `json:"format"`      // "json", "text"
	Output     string `json:"output"`      // "stdout", "stderr"
	ShowSource bool   `json:"show_source"` // показывать файл:линию
}

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := &Config{
		Server: ServerConfig{
			HttpPort:          getEnv("HTTP_PORT", "8080"),
			ReadTimeout:       time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 30)) * time.Second,
			WriteTimeout:      time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 30)) * time.Second,
			IdleTimeout:       time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 300)) * time.Second,
			ReadHeaderTimeout: time.Duration(getEnvAsInt("SERVER_READ_HEADER_TIMEOUT", 10)) * time.Second,
			MaxHeaderBytes:    getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1048576), // 1MB
			EnableKeepAlive:   getEnvAsBool("SERVER_ENABLE_KEEPALIVE", true),
			KeepAlivePeriod:   time.Duration(getEnvAsInt("SERVER_KEEPALIVE_PERIOD", 30)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "renti_kz"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		S3: S3Config{
			Region:    getEnv("AWS_REGION", "us-east-1"),
			Bucket:    getEnv("AWS_S3_BUCKET", "renti-kz"),
			AccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "access_secret"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh_secret"),
			AccessTTL:     time.Duration(getEnvAsInt("JWT_ACCESS_TTL", 24)) * time.Hour,
			RefreshTTL:    time.Duration(getEnvAsInt("JWT_REFRESH_TTL", 720)) * time.Hour,
		},
		App: AppConfig{
			PasswordSalt: getEnv("PASSWORD_SALT", "default_salt"),
			Environment:  getEnv("APP_ENV", "development"),
			Timezone:     getEnv("APP_TIMEZONE", DefaultTimezone),
		},
		Migration: MigrationConfig{
			AutoMigrate:    getEnvAsBool("AUTO_MIGRATE", true),
			MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),
		},
		Tuya: TuyaConfig{
			ClientID:     getEnv("TUYA_CLIENT_ID", ""),
			ClientSecret: getEnv("TUYA_CLIENT_SECRET", ""),
			APIBase:      getEnv("TUYA_API_BASE", "https://openapi.tuyaeu.com"),
			TimeZone:     getEnv("TUYA_TIMEZONE", "+05:00"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 100),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 50),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 1),
			DialTimeout:  time.Duration(getEnvAsInt("REDIS_DIAL_TIMEOUT", 5)) * time.Second,
			ReadTimeout:  time.Duration(getEnvAsInt("REDIS_READ_TIMEOUT", 1)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("REDIS_WRITE_TIMEOUT", 1)) * time.Second,
			PoolTimeout:  time.Duration(getEnvAsInt("REDIS_POOL_TIMEOUT", 2)) * time.Second,
		},
		Notification: NotificationConfig{
			RedisQueueName: getEnv("NOTIFICATION_QUEUE_NAME", "notification_queue"),
			PollInterval:   time.Duration(getEnvAsInt("NOTIFICATION_POLL_INTERVAL", 10)) * time.Second,
		},
		OTP: OTPConfig{
			Token:    getEnv("OTP_TOKEN", "cifpabrnvzizpgboqgitteckjitevjqx"),
			APIBase:  getEnv("OTP_API_BASE", "http://isms.center/v1/validation"),
			From:     getEnv("OTP_FROM", "KiT_Notify"),
			Template: getEnv("OTP_TEMPLATE", "Ваш код для renti.kz: [:pin]"),
		},
		FreedomPay: FreedomPayConfig{
			MerchantID: getEnv("FREEDOMPAY_MERCHANT_ID", ""),
			SecretKey:  getEnv("FREEDOMPAY_SECRET_KEY", ""),
			APIBase:    getEnv("FREEDOMPAY_API_BASE", "https://api.freedompay.kz"),
			WebhookURL: getEnv("FREEDOMPAY_WEBHOOK_URL", ""),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "debug"),
			Format:     getEnv("LOG_FORMAT", "text"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			ShowSource: getEnvAsBool("LOG_SHOW_SOURCE", false),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}
