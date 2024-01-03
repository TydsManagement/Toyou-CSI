package service

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	HostIP    string `yaml:"hostIP"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	PoolName  string `yaml:"pool"`
	StripSize int    `yaml:"strip_size"`
}

func ReadconfigFromFile(filePath string) (*Configuration, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	conf := &Configuration{}
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
