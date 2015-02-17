package main

import (
	"log"
	"os"

	"github.com/wmgaca/gostorm/drivers/memcache"
	"github.com/wmgaca/gostorm/drivers/redis"
)

const disapprovalLook string = "ಠ_ಠ"

const (
	Redis = iota
	Memcached
	MySQL
)

var Datastrores = []int{Redis, Memcached, MySQL}

// ConnectionStrings containes possible env var names for datastores connection strings
var ConnectionStrings = map[int][]string{
	Redis:     []string{"REDISTOGO_URL", "REDIS_URL"},
	Memcached: []string{"MEMCACHED_URL"},
	MySQL:     []string{"MYSQL_URL"},
}

var Drivers = map[int]func(string) (*Driver, error){
	Redis:     redis.New,
	Memcached: memcache.New,
	MySQL:     nil,
}

func GetDrivers() []Driver {
	var drivers []Driver

	// for datastoreType := range ConnectionStrings {
	// 	log.Printf("=> %d", datastoreType)
	//
	// 	driver := Drivers[datastoreType]
	// 	if driver == nil {
	// 		continue
	// 	}
	//
	// 	for _, connString := range ConnectionStrings[datastoreType] {
	// 		if value := os.Getenv(connString); len(value) > 0 {
	// 			ret := driver.(func(string) (*Driver, error))(value)
	// 			drivers = append(drivers, *ret)
	// 		}
	//
	// 	}
	//
	// }

	return []Driver{}
}

// GetRedisURL returnes Redis connection string, if present
// func GetRedisURL() (string, error) {
// 	for _, name := range RedisConnStrings {
// 		value := os.Getenv(name)
// 		if len(value) > 0 {
// 			return value, nil
// 		}
// 	}
//
// 	return "", errors.New("Redis connection string not found")
// }

// ExitWithErr prints a given error to stdout and terminates the program
func ExitWithErr(err error) {
	log.Fatal(err)
	os.Exit(-1)
}
