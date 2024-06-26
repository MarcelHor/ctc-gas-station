package main

import (
	"fmt"
	"time"
)

func printStats(gs *GasStation) {
	fmt.Println("\n====================")
	fmt.Println("SIMULATION STATISTICS")
	fmt.Println("====================")

	statsByFuelType := make(map[FuelType]*FuelStats)
	regStats := &RegisterStats{TotalCars: 0, TotalTime: 0, MaxQueueTime: 0}

	for _, car := range gs.allCars {
		fuelTypeStats, exists := statsByFuelType[car.StationType]
		if !exists {
			fuelTypeStats = &FuelStats{}
			statsByFuelType[car.StationType] = fuelTypeStats
		}

		fuelTypeStats.TotalCars++
		var serviceAndQueueTime time.Duration
		if gs.countQueueTimes {
			serviceAndQueueTime = car.ServiceTime + car.ServiceQueueTime
		} else {
			serviceAndQueueTime = car.ServiceTime
		}
		fuelTypeStats.TotalTime += serviceAndQueueTime
		if serviceAndQueueTime > fuelTypeStats.MaxQueueTime {
			fuelTypeStats.MaxQueueTime = serviceAndQueueTime
		}

		regStats.TotalCars++
		var registerTime time.Duration
		if gs.countQueueTimes {
			registerTime = car.RegisterTime + car.RegisterQueueTime
		} else {
			registerTime = car.RegisterTime
		}
		regStats.TotalTime += registerTime
		if registerTime > regStats.MaxQueueTime {
			regStats.MaxQueueTime = registerTime
		}
	}

	for fuelType, stats := range statsByFuelType {
		if stats.TotalCars > 0 {
			avgTime := stats.TotalTime / time.Duration(stats.TotalCars)
			fmt.Printf("%s - Total Cars: %d, Total Time: %s, Avg Time: %s, Max Queue Time: %s\n",
				fuelType, stats.TotalCars, stats.TotalTime, avgTime, stats.MaxQueueTime)
		}
	}

	if regStats.TotalCars > 0 {
		avgRegTime := regStats.TotalTime / time.Duration(regStats.TotalCars)
		fmt.Printf("\nRegister Stats:\nTotal Cars: %d, Total Time: %s, Avg Time: %s, Max Queue Time: %s\n",
			regStats.TotalCars, regStats.TotalTime, avgRegTime, regStats.MaxQueueTime)
	}
}
