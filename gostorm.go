package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

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

func init() {
	mySqlConnString = os.Getenv("MYSQL_CONN_STRING")
	if len(mySqlConnString) == 0 {
		ExitWithErr(errors.New("Missing MYSQL_CONN_STRING env var, are we?"))
	}

	redisConnString = os.Getenv("REDIS_CONN_STRING")
	if len(redisConnString) == 0 {
		ExitWithErr(errors.New("Missing REDIS_CONN_STRING env var, are we?"))
	}

	memcachedConnString = os.Getenv("MEMCACHED_CONN_STRING")
	if len(memcachedConnString) == 0 {
		ExitWithErr(errors.New("Missing MEMCACHED_CONN_STRING env var, are we?"))
	}

	if len(os.Getenv("DEBUG")) > 0 {
		log.Println("Debug mode on.")
		Debug = true
	}
}

// GostormConfig is Gostorm's config
type GostormConfig struct {
	redisConn     redis.Conn
	myConn        *sql.DB
	memcachedConn memcache.Client
}

// NewGostorm sets up Gostorm's connections
func NewGostorm() *GostormConfig {
	// Redis
	redisConn, err := redis.Dial(redisProtocol, redisConnString)
	if Debug {
		log.Printf("Connecting to Redis => %s", redisConnString)
	}
	if err != nil {
		log.Fatal("Can't connect to Redis.")
		redisConn = nil
	} else {
		log.Println("Connecting to Redis => OK")
	}

	// MySQL
	myConn, err := sql.Open("mysql", mySqlConnString)
	if Debug {
		log.Printf("Connecting to MySQL => %s", mySqlConnString)
	}
	if err != nil {
		log.Fatal("Can't connect to MySQL.")
		myConn = nil
	} else {
		log.Println("Connecting to MySQL => OK")
	}

	// Memcached
	if Debug {
		log.Printf("Connecting to memcached => %s", memcachedConnString)
	}
	memcachedConn := memcache.New(memcachedConnString)
	if err != nil {
		log.Fatal("Can't connect to memcached.")
		memcachedConn = nil
	} else {
		log.Println("Connecting to memcached => OK")
	}

	return &GostormConfig{
		redisConn:     redisConn,
		myConn:        myConn,
		memcachedConn: *memcachedConn,
	}
}

func (gs *GostormConfig) getFromMemcached(key string, retChan chan string, errChan chan error) {
	ret, err := gs.memcachedConn.Get("go:test")

	if err != nil {
		errChan <- err
	} else {
		retChan <- string(ret.Value[:])
	}

}

func (gs *GostormConfig) getFromRedis(key string, retChan chan string, errChan chan error) {
	if gs.redisConn == nil {
		errChan <- errors.New("Redis: connection error, something's seriously fucked.")
	} else {
		ret, err := redis.String(gs.redisConn.Do("get", key))
		if err != nil {
			errChan <- err
		} else {
			retChan <- ret
		}
	}
}

// Get a value by key
func (gs *GostormConfig) Get(key string) (string, error) {
	if gs == nil {
		return "", errors.New("Gostorm.Get: something went terribly, terribly wrong.")
	}

	retChan := make(chan string)
	errChan := make(chan error)

	go gs.getFromRedis(key, retChan, errChan)
	go gs.getFromMemcached(key, retChan, errChan)

	var (
		ret string
		err error
	)

	retCount := 0

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
		}
	}
}

func main() {
	gs := NewGostorm()

	_, err := gs.Get("go:test")
	if err != nil {
		log.Fatal(err)
		return
	}
}
