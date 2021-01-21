package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	CONFIG_LOCATION = "/etc/mailway"
)

type InstanceConfig struct {
	Hostname string `json:"hostname"`
}

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
	ServerId        string `yaml:"server_id"`
	ServerJWT       string `yaml:"server_jwt"`
	IntanceHostname string `yaml:"instance_hostname"`
}

func writeServerJWT(jwt string) error {
	file := path.Join(CONFIG_LOCATION, "conf.d", "server-jwt.yml")
	data := fmt.Sprintf("server_jwt: \"%s\"\n", jwt)
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write file")
	}

	return nil
}

func writeInstanceConfig(config InstanceConfig) error {
	file := path.Join(CONFIG_LOCATION, "conf.d", "instance.yml")
	data := ""
	data += fmt.Sprintf("instance_hostname: \"%s\"", config.Hostname)
	data += "\n"
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
