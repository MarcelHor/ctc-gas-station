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

type GasStation struct {
	stations        []*Station
	registers       []*Register
	wg              sync.WaitGroup
	stationWGs      map[FuelType]*sync.WaitGroup
	registerWG      sync.WaitGroup
	allCars         []*Car
	countQueueTimes bool
}

type Car struct {
	ID                int
	StationType       FuelType
	ArrivalAtStation  time.Time
	ServiceStartTime  time.Time
	ServiceEndTime    time.Time
	ServiceTime       time.Duration
	ServiceQueueTime  time.Duration
	ArrivalAtReg      time.Time
	RegisterStartTime time.Time
	RegisterEndTime   time.Time
	RegisterTime      time.Duration
	RegisterQueueTime time.Duration
}

type Station struct {
	ID          int
	StationType FuelType
	minServe    int
	maxServe    int
	queue       chan *Car
}

type Register struct {
	ID        int
	minHandle int
	maxHandle int
	queue     chan *Car
}

type FuelStats struct {
	TotalCars    int
	TotalTime    time.Duration
	MaxQueueTime time.Duration
}

type RegisterStats struct {
	TotalCars     int
	TotalTime     time.Duration
	MaxQueueTime  time.Duration
	TotalRegister int
}

func initGasStation(config Config) *GasStation {
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
		}
		registers = append(registers, register)
	}

	return &GasStation{
		stations:        stations,
		registers:       registers,
		stationWGs:      stationWGs,
		registerWG:      registerWG,
		countQueueTimes: false,
	}
}

// sakr v rare occasions race condition
var mu sync.Mutex

func spawnCars(gStation *GasStation, config Config) {
	for i := 0; i < config.Cars.Count; i++ {
		gStation.wg.Add(1)
		go func(carID int) {
			car := &Car{ID: carID, StationType: getRandomFuelType(), ArrivalAtStation: time.Now()}
			mu.Lock()
			gStation.allCars = append(gStation.allCars, car)
			mu.Unlock()
			station := getStationWithShortestQueue(gStation.stations, car.StationType)
			station.queue <- car

			arrivalTime := time.Duration(rand.Intn(int(config.Cars.ArrivalTimeMax.Duration.Milliseconds())-int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())+1)+int(config.Cars.ArrivalTimeMin.Duration.Milliseconds())) * time.Millisecond
			time.Sleep(arrivalTime)
		}(i)
	}
}

func stationRoutine(station *Station, gs *GasStation) {
	for car := range station.queue {
		car.ServiceStartTime = time.Now()

		serveTime := time.Duration(rand.Intn(station.maxServe-station.minServe+1)+station.minServe) * time.Millisecond
		car.ServiceTime = serveTime
		time.Sleep(serveTime)

		car.ServiceEndTime = time.Now()
		car.ServiceQueueTime = time.Since(car.ServiceStartTime) - serveTime

		register := getRegisterWithShortestQueue(gs.registers)
		car.ArrivalAtReg = time.Now()
		register.queue <- car
	}
}

func registerRoutine(register *Register, gs *GasStation) {
	for car := range register.queue {
		car.RegisterStartTime = time.Now()

		handleTime := time.Duration(rand.Intn(register.maxHandle-register.minHandle+1)+register.minHandle) * time.Millisecond
		car.RegisterTime = handleTime
		time.Sleep(handleTime)

		car.RegisterEndTime = time.Now()
		car.RegisterQueueTime = time.Since(car.RegisterStartTime) - handleTime

		if !gs.countQueueTimes {
			fmt.Println("Car", car.ID, "served in", car.RegisterTime+car.ServiceTime, "at station", car.StationType)
		} else {
			fmt.Println("Car", car.ID, "served in", car.RegisterQueueTime+car.ServiceQueueTime+car.RegisterTime+car.ServiceTime, "fuel type: ", car.StationType)
		}

		gs.wg.Done()
	}
}

func main() {
	config, err := parseConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(configToString(config))

	gs := initGasStation(config)
	gs.allCars = []*Car{}

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
