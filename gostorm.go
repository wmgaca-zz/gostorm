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

func (gs *GostormConfig) getFromMemcached(key string) (string, error) {
	ret, err := gs.memcachedConn.Get("go:test")

	return string(ret.Value[:]), err
}

func (gs *GostormConfig) getFromRedis(key string) (string, error) {
	if gs.redisConn == nil {
		return "", errors.New("Redis: connection error, something's seriously fucked.")
	}

	return redis.String(gs.redisConn.Do("get", key))
}

// Get a value by key
func (gs *GostormConfig) Get(key string) (string, error) {
	if gs == nil {
		return "", errors.New("Gostorm.Get: something went terribly, terribly wrong.")
	}

	ret, err := gs.getFromRedis(key)
	if err != nil {
		return "", err
	}

	log.Printf("Gostorm.Get(%s) => Redis     => %s", key, ret)

	ret, err = gs.getFromMemcached(key)
	if err != nil {
		return "", err
	}

	log.Printf("Gostorm.Get(%s) => Memcahced => %s", key, ret)

	return ret, nil
}

func main() {
	gs := NewGostorm()

	_, err := gs.Get("go:test")
	if err != nil {
		log.Fatal(err)
		return
	}
}
