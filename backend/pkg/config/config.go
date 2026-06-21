package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	NATS       NATSConfig
	Collectors CollectorsConfig
	Alert      AlertConfig
}

type ServerConfig struct {
	Host string
	Port int
	Mode string // debug | release
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type NATSConfig struct {
	URL string
}

type CollectorsConfig struct {
	SyslogPort    int
	SyslogAddress string
	LinuxEnabled  bool
	AgentPort     int
}

type AlertConfig struct {
	SlackWebhook string
	EmailSMTP    string
	EmailFrom    string
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "siem")
	viper.SetDefault("database.user", "siem")
	viper.SetDefault("database.password", "siem")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("nats.url", "nats://localhost:4222")

	viper.SetDefault("collectors.syslogport", 514)
	viper.SetDefault("collectors.syslogaddress", "0.0.0.0")
	viper.SetDefault("collectors.linuxenabled", true)
	viper.SetDefault("collectors.agentport", 9000)
}
