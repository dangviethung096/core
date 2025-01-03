package core

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type CoreConfig struct {
	Debug             bool               `yaml:"debug"`
	Server            ServerConfig       `yaml:"server"`
	SecureServer      SecureServerConfig `yaml:"secure_server"`
	Context           ContextConfig      `yaml:"context"`
	IdGenerator       IdGenerator        `yaml:"id_generator"`
	Database          Database           `yaml:"database"`
	SecondaryDatabase Database           `yaml:"secondary_database"`
	NatsQueue         NatsQueue          `yaml:"nats_queue"`
	Redis             RedisConfig        `yaml:"redis"`
	Proxy             ProxyConfig        `yaml:"proxy"`
	HttpClient        HttpClientConfig   `yaml:"http_client"`
	Scheduler         SchedulerConfig    `yaml:"scheduler"`
	Emqx              EmqxConfig         `yaml:"emqx"`
}

type ServerConfig struct {
	Port      int    `yaml:"port"`
	Name      string `yaml:"name"`
	CacheHtml bool   `yaml:"cache_html"`
}

type SecureServerConfig struct {
	Use       bool   `yaml:"use"`
	Port      int    `yaml:"port"`
	Name      string `yaml:"name"`
	CacheHtml bool   `yaml:"cache_html"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
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
	DBType       string `yaml:"db_type"`
}

type NatsQueue struct {
	Use bool   `yaml:"use"`
	Url string `yaml:"url"`
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

type EmqxConfig struct {
	Use          bool   `yaml:"use"`
	Broker       string `yaml:"broker"`
	PrefixClient string `yaml:"prefix_client_id"`
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
