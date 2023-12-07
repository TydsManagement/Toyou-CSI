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

// Volume represents a storage volume with various attributes.
type Volume struct {
	ID         string // Unique identifier for the volume
	Size       int64  // Size of the volume in bytes
	VolumeType string // Type of the volume (e.g., SSD, HDD)
	// Add other fields as necessary
}
