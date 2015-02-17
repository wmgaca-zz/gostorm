package redis

import (
	// "errors"

	"log"
	"net/url"
	"strings"

	redigo "github.com/garyburd/redigo/redis"
	gostorm "github.com/wmgaca/gostorm"
)

const redisProtocol = "tcp"

// Driver for Gostorm
type Driver struct {
	conn redigo.Conn
}

// New returns a new RedisDriver, duh.
func New(connString string) (*gostorm.Driver, error) {
	redisURL, err := url.Parse(connString)
	if err != nil {
		return nil, err
	}

	auth := ""
	if redisURL.User != nil {
		if password, ok := redisURL.User.Password(); ok {
			auth = password
		}
	}

	conn, err := redigo.Dial(redisProtocol, redisURL.Host)
	if err != nil {
		return nil, err
	}

	if len(auth) > 0 {
		_, err = conn.Do("AUTH", auth)
		if err != nil {
			return nil, err
		}
	}

	if len(redisURL.Path) > 1 {
		db := strings.TrimPrefix(redisURL.Path, "/")
		conn.Do("SELECT", db)
	}

	log.Println("redis.New OK")

	return &Driver{conn: conn}, nil
}

// Get return a value for a given key or an error if occured
func (drv *Driver) Get(key string, retChan chan string, errChan chan error) {
	ret, err := redigo.String(drv.conn.Do("get", key))

	if err != nil {
		errChan <- err
	} else {
		retChan <- ret
	}
}

// Set sets data :)
func (drv *Driver) Set(key, value string, retChan chan string, errChan chan error) {
	log.Printf("redis.set %s=%s", key, value)
	ret, err := redigo.String(drv.conn.Do("set", key, value))

	if err != nil {
		log.Printf("redis.set err=%s", err.Error())
		errChan <- err
	} else {
		log.Println("redis.set OK")
		retChan <- ret
	}
}
