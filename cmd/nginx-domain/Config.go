package main

import (
	"flag"
	"os"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port    string `yaml:"port"`
	} `yaml:"server"`
	path `yaml:"path"`
}
func ParseFlags() (string, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")
	flag.Parse()
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}
	return configPath, nil
}
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}