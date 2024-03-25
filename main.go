package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type FuelType string

const (
	Gas      FuelType = "gas"
	Diesel   FuelType = "diesel"
	LPG      FuelType = "lpg"
	Electric FuelType = "electric"
)

type gasStation struct {
	stations  []*Station
	registers []*Register
	wg        sync.WaitGroup
	fuelStats map[FuelType]*FuelStats
	regStats  RegisterStats
}

type Car struct {
	ID          int
	StationType FuelType
}

type Station struct {
	ID          int
	StationType FuelType
	minServe    int
	maxServe    int
	queue       chan *Car
	isBusy      bool
}

type Register struct {
	ID        int
	minHandle int
	maxHandle int
	isBusy    bool
	queue     chan *Car
}

type FuelStats struct {
	TotalCars    int
	TotalTime    time.Duration
	MaxQueueTime time.Duration
}

type RegisterStats struct {
	TotalCars    int
	TotalTime    time.Duration
	MaxQueueTime time.Duration
}

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

func initGasStation(config Config) *gasStation {
	var stations []*Station
	var registers []*Register

	for fuelType, sConf := range config.Stations {
		for i := 0; i < sConf.Count; i++ {
			station := &Station{
				ID:          i,
				StationType: fuelType,
				minServe:    sConf.ServeTimeMin,
				maxServe:    sConf.ServeTimeMax,
				queue:       make(chan *Car, config.Cars.Count),
				isBusy:      false,
			}
			stations = append(stations, station)
		}
	}

	for i := 0; i < config.Registers.Count; i++ {
		register := &Register{
			ID:        i,
			minHandle: config.Registers.HandleTimeMin,
			maxHandle: config.Registers.HandleTimeMax,
			queue:     make(chan *Car, config.Cars.Count),
			isBusy:    false,
		}
		registers = append(registers, register)
	}

	return &gasStation{
		stations:  stations,
		registers: registers,
		fuelStats: map[FuelType]*FuelStats{
			Gas:      &FuelStats{},
			Diesel:   &FuelStats{},
			LPG:      &FuelStats{},
			Electric: &FuelStats{},
		},
		regStats: RegisterStats{},
	}
}

func spawnCars(gStation *gasStation, config Config) {
	for i := 0; i < config.Cars.Count; i++ {
		car := &Car{
			ID:          i,
			StationType: getRandomFuelType(),
		}

		gStation.wg.Add(1)
		station := getStationWithShortestQueue(gStation.stations, car.StationType)
		station.queue <- car
		fmt.Printf("[%s] Car %d arrived at station queue type %s number %d\n", time.Now().Format("15:04:05"), car.ID, car.StationType, station.ID)
		arrivalTime := rand.Intn(config.Cars.ArrivalTimeMax-config.Cars.ArrivalTimeMin) + config.Cars.ArrivalTimeMin
		time.Sleep(time.Duration(arrivalTime) * time.Second)
	}
}

func stationRoutine(station *Station, gs *gasStation) {
	for car := range station.queue {
		fmt.Printf("[%s] Car %d arrived at station type %s number %d\n", time.Now().Format("15:04:05"), car.ID, car.StationType, station.ID)
		station.isBusy = true
		serveTime := rand.Intn(station.maxServe-station.minServe) + station.minServe
		time.Sleep(time.Duration(serveTime) * time.Second)
		fmt.Printf("[%s] Car %d served at station type %s number %d\n", time.Now().Format("15:04:05"), car.ID, car.StationType, station.ID)

		register := getRegisterWithShortestQueue(gs.registers)
		register.queue <- car
		station.isBusy = false
	}
}

func registerRoutine(register *Register, gs *gasStation) {
	for car := range register.queue {
		fmt.Printf("[%s] Car %d arrived at register number %d\n", time.Now().Format("15:04:05"), car.ID, register.ID)
		register.isBusy = true
		handleTime := rand.Intn(register.maxHandle-register.minHandle) + register.minHandle
		time.Sleep(time.Duration(handleTime) * time.Second)
		fmt.Printf("[%s] Car %d handled at register number %d\n", time.Now().Format("15:04:05"), car.ID, register.ID)
		gs.wg.Done()
		register.isBusy = false
	}
}
func main() {
	config, err := parseConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(configToString(config))

	gs := initGasStation(config)

	for _, station := range gs.stations {
		go stationRoutine(station, gs)
	}

	for _, register := range gs.registers {
		go registerRoutine(register, gs)
	}

	spawnCars(gs, config)

	gs.wg.Wait()

	fmt.Println("Simulation finished")
}
