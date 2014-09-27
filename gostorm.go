package main

import (
	"errors"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
)

var (
	// Debug mode, more verbose if true
	Debug = false

	mySqlConnString string

	redisConnString string

	memcachedConnString string
)

const redisProtocol = "tcp"

const defaultTimeout = 10 * time.Second

// GostormDatastoreDriver describes the interface for Gostorm's datasource driver
type GostormDatastoreDriver interface {
	// Get value from datastore
	Get(string, chan string, chan error)
}

// MemcachedDriver for Gostorm
type MemcachedDriver struct {
	conn memcache.Client
}

// NewMemcachedDriver returns a new RedisDriver, duh.
func NewMemcachedDriver() *MemcachedDriver {
	memcachedConnString = os.Getenv("MEMCACHED_CONN_STRING")
	if len(memcachedConnString) == 0 {
		ExitWithErr(errors.New("Missing MEMCACHED_CONN_STRING env var, are we?"))
	}

	if Debug {
		log.Printf("Connecting to memcached => %s", memcachedConnString)
	}

	return &MemcachedDriver{
		conn: *memcache.New(memcachedConnString),
	}
}

// Get gets data, duh
func (drv *MemcachedDriver) Get(key string, retChan chan string, errChan chan error) {
	ret, err := drv.conn.Get("go:test")

	if err != nil {
		errChan <- err
	} else {
		retChan <- string(ret.Value[:])
	}
}

// RedisDriver for Gostorm
type RedisDriver struct {
	conn redis.Conn
}

// NewRedisDriver returns a new RedisDriver, duh.
func NewRedisDriver() *RedisDriver {
	redisConnString = os.Getenv("REDIS_CONN_STRING")
	if len(redisConnString) == 0 {
		ExitWithErr(errors.New("Missing REDIS_CONN_STRING env var, are we?"))
	}

	conn, err := redis.Dial(redisProtocol, redisConnString)
	if Debug {
		log.Printf("Connecting to Redis => %s", redisConnString)
	}
	if err != nil {
		log.Fatal("Can't connect to Redis.")
		conn = nil
	} else {
		log.Println("Connecting to Redis => OK")
	}

	return &RedisDriver{
		conn: conn,
	}
}

// Get return a value for a given key or an error if occured
func (drv *RedisDriver) Get(key string, retChan chan string, errChan chan error) {
	if drv.conn == nil {
		errChan <- errors.New("Redis: connection error, something's seriously fucked.")
	} else {
		ret, err := redis.String(drv.conn.Do("get", key))
		if err != nil {
			errChan <- err
		} else {
			retChan <- ret
		}
	}
}

func init() {
	mySqlConnString = os.Getenv("MYSQL_CONN_STRING")
	if len(mySqlConnString) == 0 {
		ExitWithErr(errors.New("Missing MYSQL_CONN_STRING env var, are we?"))
	}

	if len(os.Getenv("DEBUG")) > 0 {
		log.Println("Debug mode on.")
		Debug = true
	}
}

// GostormConfig is Gostorm's config
type GostormConfig struct {
	drivers []GostormDatastoreDriver
}

// NewGostorm sets up Gostorm's connections
func NewGostorm(drivers ...GostormDatastoreDriver) *GostormConfig {

	// log.Println(drivers)
	for _, driver := range drivers {
		log.Println("A driver =>")
		log.Println(driver)
		log.Println(reflect.TypeOf(driver))
	}

	// // MySQL
	// myConn, err := sql.Open("mysql", mySqlConnString)
	// if Debug {
	// 	log.Printf("Connecting to MySQL => %s", mySqlConnString)
	// }
	// if err != nil {
	// 	log.Fatal("Can't connect to MySQL.")
	// 	myConn = nil
	// } else {
	// 	log.Println("Connecting to MySQL => OK")
	// }

	return &GostormConfig{
		drivers: drivers,
	}
}

// GetWithTimeout a value by key
func (gs *GostormConfig) GetWithTimeout(key string, timeout time.Duration) (string, error) {
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
func (gs *GostormConfig) Get(key string) (string, error) {
	return gs.GetWithTimeout(key, defaultTimeout)
}

func main() {
	redisDriver := NewRedisDriver()
	memcachedDriver := NewMemcachedDriver()

	log.Println(reflect.TypeOf(redisDriver))
	log.Println(reflect.TypeOf(memcachedDriver))

	gs := NewGostorm(redisDriver, memcachedDriver)

	_, err := gs.GetWithTimeout("go:test", 3*time.Second)
	if err != nil {
		log.Fatal(err)
		return
	}
}
