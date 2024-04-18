package main

import "math/rand"

func getRandomFuelType() FuelType {
	fuelTypes := []FuelType{Gas, Diesel, LPG, Electric}
	return fuelTypes[rand.Intn(len(fuelTypes))]
}

func getStationWithShortestQueue(stations []*Station, carType FuelType) *Station {
	var shortestQueue []*Station
	for _, station := range stations {
		if station.StationType == carType {
			if len(shortestQueue) == 0 {
				shortestQueue = append(shortestQueue, station)
			} else if len(station.queue) < len(shortestQueue[0].queue) {
				shortestQueue = []*Station{station}
			} else if len(station.queue) == len(shortestQueue[0].queue) {
				shortestQueue = append(shortestQueue, station)
			}
		}
	}
	if len(shortestQueue) == 1 {
		return shortestQueue[0]
	}
	return shortestQueue[rand.Intn(len(shortestQueue))]
}

func getRegisterWithShortestQueue(registers []*Register) *Register {
	var shortestQueue []*Register
	for _, register := range registers {
		if len(shortestQueue) == 0 {
			shortestQueue = append(shortestQueue, register)
		} else if len(register.queue) < len(shortestQueue[0].queue) {
			shortestQueue = []*Register{register}
		} else if len(register.queue) == len(shortestQueue[0].queue) {
			shortestQueue = append(shortestQueue, register)
		}
	}
	if len(shortestQueue) == 1 {
		return shortestQueue[0]
	}
	return shortestQueue[rand.Intn(len(shortestQueue))]
}
