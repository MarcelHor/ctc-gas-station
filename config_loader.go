package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type DurationConfig struct {
	Duration time.Duration
}

func (d *DurationConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return err
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

type CarConfig struct {
	Count          int            `yaml:"count"`
	ArrivalTimeMin DurationConfig `yaml:"arrival_time_min"`
	ArrivalTimeMax DurationConfig `yaml:"arrival_time_max"`
}

type StationConfig struct {
	Count        int            `yaml:"count"`
	ServeTimeMin DurationConfig `yaml:"serve_time_min"`
	ServeTimeMax DurationConfig `yaml:"serve_time_max"`
}

type RegisterConfig struct {
	Count         int            `yaml:"count"`
	HandleTimeMin DurationConfig `yaml:"handle_time_min"`
	HandleTimeMax DurationConfig `yaml:"handle_time_max"`
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

func configToString(config Config) string {
	configStr := "GAS STATION CONFIGURATION\n" +
		"=========================\n" +
		"Cars:\n" +
		fmt.Sprintf("  Count: %d\n", config.Cars.Count) +
		fmt.Sprintf("  Arrival Time Min: %v\n", config.Cars.ArrivalTimeMin.Duration) +
		fmt.Sprintf("  Arrival Time Max: %v\n", config.Cars.ArrivalTimeMax.Duration) +
		"\n" +
		"Stations:\n"
	for fuelType, stationConfig := range config.Stations {
		configStr += fmt.Sprintf("%s station has %d stations\n", fuelType, stationConfig.Count) +
			fmt.Sprintf("  Serve Time Min: %v\n", stationConfig.ServeTimeMin.Duration) +
			fmt.Sprintf("  Serve Time Max: %v\n", stationConfig.ServeTimeMax.Duration) + "\n"
	}
	configStr += "Registers:\n" +
		fmt.Sprintf("  Count: %d\n", config.Registers.Count) +
		fmt.Sprintf("  Handle Time Min: %v\n", config.Registers.HandleTimeMin.Duration) +
		fmt.Sprintf("  Handle Time Max: %v\n", config.Registers.HandleTimeMax.Duration) +
		"=========================\n"
	return configStr
}
