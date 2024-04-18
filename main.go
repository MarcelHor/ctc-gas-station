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
	ID                int
	StationType       FuelType
	ArrivalAtStation  time.Time
	ArrivalAtReg      time.Time
	ServiceQueueTime  time.Duration
	ServiceTime       time.Duration
	RegisterQueueTime time.Duration
	RegisterTime      time.Duration
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
			ID:               i,
			StationType:      getRandomFuelType(),
			ArrivalAtStation: time.Now(),
		}

		gStation.wg.Add(1)
		station := getStationWithShortestQueue(gStation.stations, car.StationType)
		station.queue <- car
		arrivalTime := rand.Intn(int(config.Cars.ArrivalTimeMax.Duration.Milliseconds())-int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())+1) + int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())
		time.Sleep(time.Duration(arrivalTime) * time.Millisecond)

	}
}

func stationRoutine(station *Station, gs *gasStation) {
	for car := range station.queue {
		startServiceTime := time.Now()
		queueTime := startServiceTime.Sub(car.ArrivalAtStation) // Výpočet čekací doby

		// Simulace obsluhy
		serveTime := time.Duration(rand.Intn(station.maxServe-station.minServe+1)+station.minServe) * time.Millisecond
		time.Sleep(serveTime)

		serviceDuration := time.Since(startServiceTime)
		totalStationTime := queueTime + serviceDuration // Zahrnutí čekací doby do celkového času
		gs.fuelStats[car.StationType].TotalCars++
		gs.fuelStats[car.StationType].TotalTime += totalStationTime
		if totalStationTime > gs.fuelStats[car.StationType].MaxQueueTime {
			gs.fuelStats[car.StationType].MaxQueueTime = totalStationTime
		}

		// Přesun do registru
		register := getRegisterWithShortestQueue(gs.registers)
		car.ArrivalAtReg = time.Now()
		register.queue <- car
	}
}

func registerRoutine(register *Register, gs *gasStation) {
	for car := range register.queue {
		startServiceTime := time.Now()
		queueTime := startServiceTime.Sub(car.ArrivalAtReg)

		handleTime := rand.Intn(register.maxHandle-register.minHandle+1) + register.minHandle
		time.Sleep(time.Duration(handleTime) * time.Millisecond)

		serviceDuration := time.Since(startServiceTime)
		totalRegisterTime := queueTime + serviceDuration
		gs.regStats.TotalCars++
		gs.regStats.TotalTime += totalRegisterTime

		if totalRegisterTime > gs.regStats.MaxQueueTime {
			gs.regStats.MaxQueueTime = totalRegisterTime
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
