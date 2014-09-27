package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/wmgaca/gostorm/drivers/memcache"
	"github.com/wmgaca/gostorm/drivers/redis"
)

// Debug mode, more verbose if true
var Debug = false

const defaultTimeout = 10 * time.Second

func init() {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Println("Debug mode on.")
		Debug = true
	}
}

// Gostorm is Gostorm's config
type Gostorm struct {
	drivers []Driver
}

// New sets up Gostorm's connections
func New(drivers ...Driver) *Gostorm {
	return &Gostorm{
		drivers: drivers,
	}
}

// GetWithTimeout a value by key
func (gs *Gostorm) GetWithTimeout(key string, timeout time.Duration) (string, error) {
	if gs == nil {
		return "", errors.New("Gostorm.Get: something went terribly, terribly wrong.")
	}

	retChan := make(chan string)
	errChan := make(chan error)

	for _, driver := range gs.drivers {
		go driver.Get(key, retChan, errChan)
	}

	var (
		ret string
		err error
	)

	retCount := 0
	startTime := time.Now()

	for {
		select {
		case ret = <-retChan:
			retCount++
			log.Printf("Got a result => %s", ret)
			return ret, nil
		case err = <-errChan:
			retCount++
			log.Printf("Got an error => %s", err)
			// 2 == number of data stores we're using at the moment ;)
			if retCount == 2 {
				return "", err
			}
		default:
			if time.Since(startTime) > timeout {
				return "", errors.New("Gostorm connection timeout.")
			}
		}
	}
}

// Get a value by key
func (gs *Gostorm) Get(key string) (string, error) {
	return gs.GetWithTimeout(key, defaultTimeout)
}

func main() {
	redisDriver, err := redis.New()
	memcachedDriver, err := memcache.New()

	drivers := []Driver{redisDriver, memcachedDriver}

	gs := New(drivers...)

	_, err = gs.GetWithTimeout("go:test", 3*time.Second)
	if err != nil {
		log.Fatal(err)
		return
	}
}
