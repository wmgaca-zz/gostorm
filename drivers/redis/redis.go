package redis

import (
	"errors"
	"log"

	redigo "github.com/garyburd/redigo/redis"
)

const redisProtocol = "tcp"

// Driver for Gostorm
type Driver struct {
	conn redigo.Conn
}

// New returns a new RedisDriver, duh.
func New(connString string) (*Driver, error) {
	conn, err := redigo.Dial(redisProtocol, connString)

	log.Printf("Connecting to Redis => %s", connString)

	if err != nil {
		return nil, errors.New("Can't connect to Redis.")
	}

	log.Println("Connecting to Redis => OK")

	return &Driver{conn: conn}, nil
}

// Get return a value for a given key or an error if occured
func (drv *Driver) Get(key string, retChan chan string, errChan chan error) {
	if drv.conn == nil {
		errChan <- errors.New("Redis: connection error, something's seriously fucked.")
	} else {
		ret, err := redigo.String(drv.conn.Do("get", key))
		if err != nil {
			errChan <- err
		} else {
			retChan <- ret
		}
	}
}
