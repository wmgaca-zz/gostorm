package memcache

import (
	"errors"
	"log"
	"os"

	gomemcache "github.com/bradfitz/gomemcache/memcache"
)

var memcachedConnString string

// Driver for Gostorm
type Driver struct {
	conn gomemcache.Client
}

// New returns a new memcache.Driver
func New() (*Driver, error) {
	memcachedConnString = os.Getenv("MEMCACHED_CONN_STRING")
	if len(memcachedConnString) == 0 {
		return nil, errors.New("Missing MEMCACHED_CONN_STRING env var, are we?")
	}

	log.Printf("Connecting to memcached => %s", memcachedConnString)

	driver := &Driver{
		conn: *gomemcache.New(memcachedConnString),
	}

	return driver, nil
}

// Get gets data ;)
func (drv *Driver) Get(key string, retChan chan string, errChan chan error) {
	ret, err := drv.conn.Get("go:test")

	if err != nil {
		errChan <- err
	} else {
		retChan <- string(ret.Value[:])
	}
}
