package main

import (
	"fmt"
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

func configToString(config Config) string {
	configStr := "GAS STATION CONFIGURATION\n" +
		"=========================\n" +
		"Cars:\n" +
		fmt.Sprintf("  Count: %d\n", config.Cars.Count) +
		fmt.Sprintf("  Arrival Time Min: %d\n", config.Cars.ArrivalTimeMin) +
		fmt.Sprintf("  Arrival Time Max: %d\n", config.Cars.ArrivalTimeMax) +
		"\n" +
		"Stations:\n"
	for fuelType, stationConfig := range config.Stations {
		configStr += fmt.Sprintf("%s station has %d stations\n", fuelType, stationConfig.Count) +
			fmt.Sprintf("  Serve Time Min: %d\n", stationConfig.ServeTimeMin) +
			fmt.Sprintf("  Serve Time Max: %d\n", stationConfig.ServeTimeMax) + "\n"
	}
	configStr += "Registers:\n" +
		fmt.Sprintf("  Count: %d\n", config.Registers.Count) +
		fmt.Sprintf("  Handle Time Min: %d\n", config.Registers.HandleTimeMin) +
		fmt.Sprintf("  Handle Time Max: %d\n", config.Registers.HandleTimeMax) +
		"=========================\n"
	return configStr
}
