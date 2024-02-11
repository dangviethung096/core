package core

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type CoreConfig struct {
	Debug       bool             `yaml:"debug"`
	Server      ServerConfig     `yaml:"server"`
	Context     ContextConfig    `yaml:"context"`
	IdGenerator IdGenerator      `yaml:"id_generator"`
	Database    Database         `yaml:"database"`
	RabbitMQ    RabbitMQConfig   `yaml:"rabbitmq"`
	Redis       RedisConfig      `yaml:"redis"`
	Proxy       ProxyConfig      `yaml:"proxy"`
	HttpClient  HttpClientConfig `yaml:"http_client"`
	Scheduler   SchedulerConfig  `yaml:"scheduler"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	Name      string `yaml:"name"`
	CacheHtml bool   `yaml:"cache_html"`
}

type ContextConfig struct {
	Timeout int `yaml:"timeout"`
}

func (contextConfig ContextConfig) GetTimeout() time.Duration {
	return time.Duration(contextConfig.Timeout) * time.Second
}

/*
* Get task timeout from config
 */
func (config CoreConfig) GetTaskTimeout() time.Duration {
	timeout := time.Second * 120
	if config.Scheduler.TaskTimeout != 0 {
		timeout = time.Second * time.Duration(config.Scheduler.TaskTimeout)
	}
	return timeout
}

/*
* Get context timeout from core config
* @return: timeout value from core config
 */
func (config CoreConfig) GetContextTimeout() time.Duration {
	return time.Duration(config.Context.Timeout) * time.Second
}

type IdGenerator struct {
	Distributed bool `yaml:"distributed"`
}

type Database struct {
	Use          bool   `yaml:"use"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"user"`
	Password     string `yaml:"pass"`
	DatabaseName string `yaml:"name"`
}

type RabbitMQConfig struct {
	Use           bool   `yaml:"use"`
	AMQPServerURL string `yaml:"amqp_server_url"`
	RetryTime     int    `yaml:"retry_time"`
}

type RedisConfig struct {
	Use  bool   `yaml:"use"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type ProxyConfig struct {
	Url string `yaml:"url"`
}

func (proxyConfig ProxyConfig) GetConfigUrl() string {
	return proxyConfig.Url
}

type HttpClientConfig struct {
	RetryTimes int `yaml:"retry_times"`
	WaitTimes  int `yaml:"wait_times"`
}

type SchedulerConfig struct {
	Use                 bool `yaml:"use"`
	TaskDoingExpiration int  `yaml:"task_doing_expiration"`
	Delay               int  `yaml:"delay"`
	Interval            int  `yaml:"interval"`
	BucketSize          int  `yaml:"bucket_size"`
	TaskTimeout         int  `yaml:"task_timeout"`
}

func loadConfigFile(configFile string) CoreConfig {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error when read config file: %s", err.Error())
	}

	// Unmarshal the YAML data into a Config struct
	var config CoreConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	return config
}
