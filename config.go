package config

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	ROOT_LOCATION   = "/etc/mailway"
	CONFIG_LOCATION = ROOT_LOCATION + "/conf.d"
)

var (
	CurrConfig *Config
)

func PrettyPrint() ([]byte, error) {
	s, err := yaml.Marshal(&CurrConfig)
	if err != nil {
		return []byte(""), err
	}
	return s, nil
}

func readAll() ([]byte, error) {
	config := []byte{}

	files, err := ioutil.ReadDir(CONFIG_LOCATION)
	if err != nil {
		return config, err
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".yml" || ext == ".yaml" {
			absFile := path.Join(CONFIG_LOCATION, file.Name())
			content, err := ioutil.ReadFile(absFile)
			if err != nil {
				return config, err
			}
			config = append(config, content...)
		}
	}

	return config, err
}

type Config struct {
	LogLevel  string `yaml:"log_level"`
	LogFormat string `yaml:"log_format"`

	ServerId         string `yaml:"server_id"`
	ServerJWT        string `yaml:"server_jwt"`
	InstanceHostname string `yaml:"instance_hostname"`
	InstanceEmail    string `yaml:"instance_email"`
	InstanceMode     string `yaml:"instance_mode"`

	PortAuth           int `yaml:"port_auth"`
	PortForwarding     int `yaml:"port_forwarding"`
	PortMaildb         int `yaml:"port_maildb"`
	PortMailout        int `yaml:"port_mailout"`
	PortWebhook        int `yaml:"port_webhook"`
	PortFrontlineSMTP  int `yaml:"port_frontline_smtp"`
	PortFrontlineSMTPS int `yaml:"port_frontline_smtps"`

	OutSMTPHost     string `yaml:"out_smtp_host"`
	OutSMTPUsername string `yaml:"out_smtp_username"`
	OutSMTPPassword string `yaml:"out_smtp_password"`
	OutSMTPPort     int    `yaml:"out_smtp_port"`
	OutDKIMPath     string `yaml:"out_dkim_path"`

	LogFrontlineError      string `yaml:"log_frontline_error"`
	LogFrontlineHTTPAccess string `yaml:"log_frontline_http_access"`
	LogFrontlineHTTPError  string `yaml:"log_frontline_http_error"`

	ForwardingLoopDetectionCount int `yaml:"forwarding_loop_detection_count"`
	ForwardingRateLimitingCount  int `yaml:"forwarding_rate_limiting_count"`

	MaildbPath string `yaml:"maildb_db_path"`

	SpamFilter bool `yaml:"spam_filter"`
}

func WriteServerJWT(jwt string) error {
	file := path.Join(CONFIG_LOCATION, "server-jwt.yml")
	data := fmt.Sprintf("server_jwt: \"%s\"\n", jwt)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	CurrConfig.ServerJWT = jwt

	return nil
}

func WriteDKIM(keyPath string) error {
	file := path.Join(CONFIG_LOCATION, "dkim.yml")
	data := fmt.Sprintf("out_dkim_path: \"%s\"\n", keyPath)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	CurrConfig.OutDKIMPath = keyPath

	return nil
}

func WriteInstanceConfig(mode, hostname, email string) error {
	file := path.Join(CONFIG_LOCATION, "instance.yml")
	data := ""
	data += fmt.Sprintf("instance_mode: \"%s\"\n", mode)
	data += fmt.Sprintf("instance_hostname: \"%s\"\n", hostname)
	data += fmt.Sprintf("instance_email: \"%s\"\n", email)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}
	CurrConfig.InstanceMode = mode
	CurrConfig.InstanceHostname = hostname
	CurrConfig.InstanceEmail = email

	return nil
}

func loadConfig() error {
	data, err := readAll()
	if err != nil {
		return errors.Wrap(err, "could not read config")
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return errors.Wrap(err, "failed to parse")
	}

	*CurrConfig = config
	return nil
}

func Init() error {
	CurrConfig = new(Config)
	if err := loadConfig(); err != nil {
		return errors.Wrap(err, "failed to load config")
	}
	log.SetLevel(CurrConfig.GetLogLevel())
	log.SetFormatter(CurrConfig.GetLogFormat())

	go func() {
		if err := watchConfig(); err != nil {
			log.Errorf("could not watch config: %s", err)
		}
	}()

	return nil
}

func watchConfig() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "could not create new watcher")
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					break
				}
				log.Debugf("%s detected config change; reloading config", event.String())
				if err := loadConfig(); err != nil {
					log.Errorf("could not load config: %s", err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					break
				}
				log.Errorf("error while watching files: %s", err)
			}
		}
	}()

	log.Debugf("start watching %s for changes", CONFIG_LOCATION)
	if err := watcher.Add(CONFIG_LOCATION); err != nil {
		return errors.Wrap(err, "failed to watch config")
	}
	<-done
	return nil
}

func (c *Config) GetLogLevel() log.Level {
	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}
	switch c.LogLevel {
	case "INFO":
		return log.InfoLevel
	case "DEBUG":
		return log.DebugLevel
	case "WARN":
		return log.WarnLevel
	}
	log.Fatalf("unknown log level: '%s'", c.LogLevel)
	panic("unreachable")
}

func (c *Config) GetLogFormat() log.Formatter {
	if c.LogFormat == "" {
		c.LogFormat = "text"
	}
	switch c.LogFormat {
	case "text":
		return &log.TextFormatter{}
	case "json":
		return &log.JSONFormatter{}
	}
	log.Fatalf("unknown log format: '%s'", c.LogFormat)
	panic("unreachable")
}

func (c *Config) IsInstanceLocal() bool {
	return c.InstanceMode == "local"
}
