package service

import (
	"os"

	"gopkg.in/yaml.v2"
)

type configuration struct {
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func ReadconfigFromFile(filePath string) (*configuration, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	conf := &configuration{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func convstr(s string) string {
	if s == "" {
		return "NULL"
	}
	return s
}
