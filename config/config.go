package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/viper"
)

const (
	configPathEnv = "PGNEXTGENCONSUMER"
)

// Config contains all config details stucture
type Config struct {
	Debug          bool           `yaml:"debug"`
	Port           string         `yaml:"port"`
	Profile        string         `yaml:"profile"`
	MySqlConfig    MySqlConfig    `yaml:"mysqlConfig"`
	LogConfig      LogConfig      `yaml:"logConfig"`
	ConsumerConfig ConsumerConfig `yaml:"consumerConfig"`
	RedisConfig    RedisConfig    `yaml:"redisConfig"`
}

type MySqlConfig struct {
	Secured            bool   `yaml:"secured"`
	Hostname           string `yaml:"hostname"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	MaxConnections     int    `yaml:"maxConn"`
	MaxIdleConnections int    `yaml:"maxIdleConn"`
	MaxConnLifetime    int    `yaml:"maxConnLifetimeMin"`
}

type RedisConfig struct {
	Secured   bool   `yaml:"secured"`
	Hostname  string `yaml:"hostname"`
	Port      string `yaml:"port"`
	QueueName string `yaml:"queueName"`
}

type ConsumerConfig struct {
	PrefetchLimit   int64 `yaml:"prefetchLimit"`
	PollDuration    int   `yaml:"pollDuration"`
	NumConsumers    int   `yaml:"numConsumers"`
	ReportBatchSize int   `yaml:"reportBatchSize"`
	ShouldLog       bool  `yaml:"shouldLog"`
}

func LoadConfig() *Config {
	configPath, _ := os.LookupEnv(configPathEnv)
	//configPath := "./config.yml"
	if configPath == "" {
		log.Fatalln("Configuration file path not found in", configPath)
	}
	log.Println("Loading configuration from", configPath)

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Cannot read config file:", err)
	}
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalln("Cannot decode config file:", err)
	}

	// mysql config
	dbByte, _ := json.Marshal(viper.GetStringMap("mysqlConfig"))
	_ = json.Unmarshal(dbByte, &config.MySqlConfig)
	if config.MySqlConfig.Secured {
		config.MySqlConfig.Password = os.Getenv("MYSQL_PASS")
		if config.MySqlConfig.Password == "" {
			panic("Mysql password not set in environment variable MYSQL_PASS, can't start service")
		}
	}

	//Consumer config
	consumerConfigBytes, _ := json.Marshal(viper.GetStringMap("consumerConfig"))
	_ = json.Unmarshal(consumerConfigBytes, &config.ConsumerConfig)

	//log config
	logConfigBytes, _ := json.Marshal(viper.GetStringMap("logConfig"))
	_ = json.Unmarshal(logConfigBytes, &config.LogConfig)

	//log config
	redisConfigBytes, _ := json.Marshal(viper.GetStringMap("redisConfig"))
	_ = json.Unmarshal(redisConfigBytes, &config.LogConfig)

	return config
}
