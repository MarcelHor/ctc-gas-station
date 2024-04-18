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
	stations   []*Station
	registers  []*Register
	wg         sync.WaitGroup
	stationWGs map[FuelType]*sync.WaitGroup
	registerWG sync.WaitGroup
	fuelStats  map[FuelType]*FuelStats
	regStats   RegisterStats
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
	stationWGs := make(map[FuelType]*sync.WaitGroup)

	for fuelType, sConf := range config.Stations {
		stationWGs[fuelType] = &sync.WaitGroup{}
		for i := 0; i < sConf.Count; i++ {
			stationWGs[fuelType].Add(1)
			station := &Station{
				ID:          i,
				StationType: fuelType,
				minServe:    int(sConf.ServeTimeMin.Duration.Milliseconds()),
				maxServe:    int(sConf.ServeTimeMax.Duration.Milliseconds()),
				queue:       make(chan *Car, 20),
				isBusy:      false,
			}
			stations = append(stations, station)
		}
	}

	var registerWG sync.WaitGroup
	for i := 0; i < config.Registers.Count; i++ {
		registerWG.Add(1)
		register := &Register{
			ID:        i,
			minHandle: int(config.Registers.HandleTimeMin.Duration.Milliseconds()),
			maxHandle: int(config.Registers.HandleTimeMax.Duration.Milliseconds()),
			queue:     make(chan *Car, 20),
			isBusy:    false,
		}
		registers = append(registers, register)
	}

	return &gasStation{
		stations:   stations,
		registers:  registers,
		stationWGs: stationWGs,
		registerWG: registerWG,
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
		arrivalTime := rand.Intn(int(config.Cars.ArrivalTimeMax.Duration.Milliseconds())-int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())) + int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())
		time.Sleep(time.Duration(arrivalTime) * time.Millisecond)
	}
}

func stationRoutine(station *Station, gs *gasStation) {
	for car := range station.queue {
		startTime := time.Now()
		fmt.Printf("[%s] Car %d processing at station type %s number %d\n", startTime.Format("15:04:05"), car.ID, car.StationType, station.ID)

		// Simulace obsluhy
		serveTime := time.Duration(rand.Intn(station.maxServe-station.minServe)+station.minServe) * time.Millisecond
		time.Sleep(serveTime)

		elapsed := time.Since(startTime)
		gs.fuelStats[car.StationType].TotalCars++
		gs.fuelStats[car.StationType].TotalTime += elapsed
		if elapsed > gs.fuelStats[car.StationType].MaxQueueTime {
			gs.fuelStats[car.StationType].MaxQueueTime = elapsed
		}

		fmt.Printf("[%s] Car %d served at station type %s number %d\n", time.Now().Format("15:04:05"), car.ID, car.StationType, station.ID)
		// Přesun do registru
		register := getRegisterWithShortestQueue(gs.registers)
		register.queue <- car
	}
}

func registerRoutine(register *Register, gs *gasStation) {
	for car := range register.queue {
		startTime := time.Now()
		fmt.Printf("[%s] Car %d arrived at register number %d\n", time.Now().Format("15:04:05"), car.ID, register.ID)
		register.isBusy = true

		handleTime := rand.Intn(register.maxHandle-register.minHandle) + register.minHandle
		time.Sleep(time.Duration(handleTime) * time.Millisecond)
		fmt.Printf("[%s] Car %d handled at register number %d\n", time.Now().Format("15:04:05"), car.ID, register.ID)

		elapsed := time.Since(startTime)
		gs.regStats.TotalCars++
		gs.regStats.TotalTime += elapsed
		if elapsed > gs.regStats.MaxQueueTime {
			gs.regStats.MaxQueueTime = elapsed
		}

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

	gs.fuelStats = make(map[FuelType]*FuelStats)
	for _, ft := range []FuelType{Gas, Diesel, LPG, Electric} {
		gs.fuelStats[ft] = &FuelStats{}
	}
	gs.regStats = RegisterStats{}

	for _, station := range gs.stations {
		go stationRoutine(station, gs)
	}

	for _, register := range gs.registers {
		go registerRoutine(register, gs)
	}

	spawnCars(gs, config)

	gs.wg.Wait()

	for _, station := range gs.stations {
		close(station.queue)
	}

	for _, register := range gs.registers {
		close(register.queue)
	}

	fmt.Println("Simulation finished")

	printStats(gs)
}

func printStats(gs *gasStation) {
	fmt.Println("\n====================")
	fmt.Println("SIMULATION STATISTICS")
	fmt.Println("====================")

	for fuelType, stats := range gs.fuelStats {
		if stats.TotalCars > 0 {
			avgTime := stats.TotalTime / time.Duration(stats.TotalCars)
			fmt.Printf("%s - Total Cars: %d, Total Time: %s, Avg Time: %s, Max Queue Time: %s\n", fuelType, stats.TotalCars, stats.TotalTime, avgTime, stats.MaxQueueTime)
		}
	}

	fmt.Println("\nRegister Stats:")
	if gs.regStats.TotalCars > 0 {
		avgTime := gs.regStats.TotalTime / time.Duration(gs.regStats.TotalCars)
		fmt.Printf("Total Cars: %d, Total Time: %s, Avg Time: %s, Max Queue Time: %s\n", gs.regStats.TotalCars, gs.regStats.TotalTime, avgTime, gs.regStats.MaxQueueTime)
	}
}
