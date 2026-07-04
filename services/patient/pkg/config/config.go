package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config holds all service configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Logging  LoggingConfig
}

// ServerConfig holds gRPC server configuration.
type ServerConfig struct {
	Port            int
	HealthCheckPort int
	GracefulTimeout int // seconds
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
}

// KafkaConfig holds Kafka configuration.
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level string // debug, info, warn, error
}

// Load loads configuration from environment variables and config file.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	// Set defaults
	v.SetDefault("server.port", 50051)
	v.SetDefault("server.healthCheckPort", 8081)
	v.SetDefault("server.gracefulTimeout", 30)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.maxConns", 25)
	v.SetDefault("logging.level", "info")

	// Try to load config file, but don't fail if it doesn't exist
	_ = v.ReadInConfig()

	// Override with environment variables
	_ = v.BindEnv("server.port", "GRPC_PORT")
	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.user", "DB_USER")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.dbname", "DB_NAME")
	_ = v.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	_ = v.BindEnv("logging.level", "LOG_LEVEL")

	cfg := &Config{
		Server: ServerConfig{
			Port:            v.GetInt("server.port"),
			HealthCheckPort: v.GetInt("server.healthCheckPort"),
			GracefulTimeout: v.GetInt("server.gracefulTimeout"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("database.host"),
			Port:     v.GetInt("database.port"),
			User:     getEnv("DB_USER", v.GetString("database.user")),
			Password: getEnv("DB_PASSWORD", v.GetString("database.password")),
			DBName:   getEnv("DB_NAME", v.GetString("database.dbname")),
			SSLMode:  v.GetString("database.sslmode"),
			MaxConns: v.GetInt("database.maxConns"),
		},
		Kafka: KafkaConfig{
			Brokers: parseBrokers(getEnv("KAFKA_BROKERS", "localhost:9092")),
			GroupID: getEnv("KAFKA_GROUP_ID", "his-patient-service"),
		},
		Logging: LoggingConfig{
			Level: v.GetString("logging.level"),
		},
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func parseBrokers(brokerStr string) []string {
	if brokerStr == "" {
		return []string{"localhost:9092"}
	}
	// For simplicity, assume comma-separated list
	// In production, use a proper parser
	return []string{brokerStr}
}
