package main

import (
	"gopkg.in/yaml.v2"
	"os"
)

type CarConfig struct {
	Count          int `yaml:"count"`
	ArrivalTimeMin int `yaml:"arrival_time_min"`
	ArrivalTimeMax int `yaml:"arrival_time_max"`
}

type StationConfig struct {
	Count        int `yaml:"count"`
	ServeTimeMin int `yaml:"serve_time_min"`
	ServeTimeMax int `yaml:"serve_time_max"`
}

type RegisterConfig struct {
	Count         int `yaml:"count"`
	HandleTimeMin int `yaml:"handle_time_min"`
	HandleTimeMax int `yaml:"handle_time_max"`
}

type Config struct {
	Cars      CarConfig                  `yaml:"cars"`
	Stations  map[FuelType]StationConfig `yaml:"stations"`
	Registers RegisterConfig             `yaml:"registers"`
}

func parseConfig(filename string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
