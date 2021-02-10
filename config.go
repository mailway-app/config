package config

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	CONFIG_LOCATION = "/etc/mailway"
)

func PrettyPrint() ([]byte, error) {
	c, err := Read()
	if err != nil {
		return []byte(""), err
	}
	s, err := yaml.Marshal(&c)
	if err != nil {
		return []byte(""), err
	}
	return s, nil
}

func readAll() ([]byte, error) {
	config := []byte{}

	confd := path.Join(CONFIG_LOCATION, "conf.d")
	files, err := ioutil.ReadDir(confd)
	if err != nil {
		return config, err
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(path.Join(confd, file.Name()))
		if err != nil {
			return config, err
		}
		config = append(config, content...)
	}

	return config, err
}

type Config struct {
	LogLevel string `yaml:"log_level"`

	ServerId         string `yaml:"server_id"`
	ServerJWT        string `yaml:"server_jwt"`
	InstanceHostname string `yaml:"instance_hostname"`
	InstanceEmail    string `yaml:"instance_email"`

	PortAuth       int `yaml:"port_auth"`
	PortForwarding int `yaml:"port_forwarding"`
	PortMaildb     int `yaml:"port_maildb"`
	PortMailout    int `yaml:"port_mailout"`

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

	SpamFilter bool `yaml:"spam_filter"`
}

func WriteServerJWT(jwt string) error {
	file := path.Join(CONFIG_LOCATION, "conf.d", "server-jwt.yml")
	data := fmt.Sprintf("server_jwt: \"%s\"\n", jwt)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	return nil
}

func WriteDKIM(keyPath string) error {
	file := path.Join(CONFIG_LOCATION, "conf.d", "dkim.yml")
	data := fmt.Sprintf("out_dkim_path: \"%s\"\n", keyPath)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	return nil
}

func WriteInstanceConfig(hostname, email string) error {
	file := path.Join(CONFIG_LOCATION, "conf.d", "instance.yml")
	data := ""
	data += fmt.Sprintf("instance_hostname: \"%s\"\n", hostname)
	data += fmt.Sprintf("instance_email: \"%s\"\n", email)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	return nil
}

func Read() (*Config, error) {
	data, err := readAll()
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse")
	}

	return &config, nil
}

func (c *Config) GetLogLevel() log.Level {
	switch c.LogLevel {
	case "":
	case "INFO":
		return log.InfoLevel
	case "DEBUG":
		return log.DebugLevel
	case "WARN":
		return log.WarnLevel
	}
	log.Fatalf("uknown log level: %s", c.LogLevel)
	panic("unreachable")
}
