package memcache

import (
	"log"

	gomemcache "github.com/bradfitz/gomemcache/memcache"
)

// Driver for Gostorm
type Driver struct {
	conn gomemcache.Client
}

// New returns a new memcache.Driver
func New(connString string) (*Driver, error) {
	log.Printf("Connecting to memcached => %s", connString)

	driver := &Driver{
		conn: *gomemcache.New(connString),
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
