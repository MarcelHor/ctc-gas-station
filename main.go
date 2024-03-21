package main

import (
	"fmt"
)

type FuelType string

const (
	Gas      FuelType = "gas"
	Diesel   FuelType = "diesel"
	LPG      FuelType = "lpg"
	Electric FuelType = "electric"
)

type Car struct {
	ID          int
	StationType FuelType
}

type Station struct {
	ID          int
	StationType FuelType
}

type Register struct {
	ID int
}

func main() {
	config, err := parseConfig("config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed Configuration: %+v\n", config)
}
