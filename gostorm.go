package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/wmgaca/gostorm/drivers/redis"
)

const (
	// DefaultTimeout for Gostorm to wait for a datastore,
	// will return ErrTimeout if the operation takes longer
	DefaultTimeout = 10 * time.Second
)

var (
	gostormInstance Gostorm

	missingEnvVarErrString = "Missing $%s environmental variable, aren't we?"

	// ErrTimeout is returned when the datastors fail
	// to return an answer in a given time
	ErrTimeout = errors.New("Gostorm query timed out.")

	// Debug mode, more verbose if true
	Debug = false

	// ServerAddr that gostorm listen on
	ServerAddr string
)

func init() {
	if len(os.Getenv("DEBUG")) > 0 {
		Debug = true
	}
	log.Printf("gostorm Debug=%t", Debug)

	ServerAddr = ":" + os.Getenv("PORT")
	if !(len(ServerAddr) > 1) {
		log.Printf(missingEnvVarErrString, "PORT")
	}
	log.Printf("gostorm ServerAddr=%s", ServerAddr)
}

// Gostorm is Gostorm, duh
type Gostorm struct {
	drivers []Driver
}

// New returns a new Gostorm instance
func New(drivers ...Driver) *Gostorm {
	return &Gostorm{
		drivers: drivers,
	}
}

// GetWithTimeout will try to fetch the value for a given key from the datastores
func (gs *Gostorm) GetWithTimeout(key string, timeout time.Duration) (string, error) {
	retChan := make(chan string)
	errChan := make(chan error)

	for _, driver := range gs.drivers {
		go driver.Get(key, retChan, errChan)
	}

	var (
		ret       string
		err       error
		retCount  = 0
		startTime = time.Now()
	)

	for {
		select {
		case ret = <-retChan:
			retCount++
			log.Printf("gostorm.ret => %s", ret)
			return ret, nil
		case err = <-errChan:
			retCount++
			log.Printf("gostorm.err => %s", err)
			if retCount == len(gs.drivers) {
				// Don't expect any more answers at this point
				return "", err
			}
		default:
			if time.Since(startTime) > timeout {
				return "", ErrTimeout
			}
		}
	}
}

// SetWithTimeout a value by key
func (gs *Gostorm) SetWithTimeout(key, value string, timeout time.Duration) error {
	retChan := make(chan string)
	errChan := make(chan error)

	for _, driver := range gs.drivers {
		go driver.Set(key, value, retChan, errChan)
	}

	var (
		ret       string
		err       error
		retCount  = 0
		startTime = time.Now()
	)

	for {
		select {
		case ret = <-retChan:
			retCount++
			log.Printf("gostorm.set => %s", ret)
			return nil
		case err = <-errChan:
			retCount++
			log.Printf("gostorm.set => %s", err)
			if retCount == len(gs.drivers) {
				// Don't expect any more answers at this point
				return err
			}
		default:
			if time.Since(startTime) > timeout {
				return ErrTimeout
			}
		}
	}
}

// Get a value for given key, uses DefaultTimeout
func (gs *Gostorm) Get(key string) (string, error) {
	return gs.GetWithTimeout(key, DefaultTimeout)
}

// Set a value for given key, uses DefaultTimeout
func (gs *Gostorm) Set(key, value string) error {
	return gs.SetWithTimeout(key, value, DefaultTimeout)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ret := "Go, baby, go!"
	log.Printf("%s / => %s", r.Method, ret)
	fmt.Fprintf(w, "%s\n", ret)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	ret, err := gostormInstance.Get(key)
	if err != nil {
		ret = err.Error()
	}

	log.Printf("%s /get/%s/ => %s", r.Method, key, ret)

	fmt.Fprintf(w, "%s\n", ret)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err   = r.ParseForm()
		key   = "?"
		value = "?"
		ret   = "?"
	)

	if err != nil {
		ret = err.Error()
	} else {
		if len(r.PostForm["key"]) == 0 || len(r.PostForm["value"]) == 0 {
			ret = "Missing key or value?"
		} else {
			key = r.PostForm["key"][0]
			value = r.PostForm["value"][0]

			ret = "SUCCESS"
			err = gostormInstance.Set(key, value)
			if err != nil {
				ret = err.Error()
			}
		}
	}

	log.Printf("%s /set/%s/ => %s", r.Method, key, ret)

	fmt.Fprintf(w, "%s\n", ret)
}

func configureRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/get/{key:[a-zA-Z0-9:.]+}/", getHandler).Methods("GET")
	router.HandleFunc("/set/", setHandler).Methods("POST")

	return router
}

func main() {
	GetDrivers()

	log.Println("gostorm.main")

	var drivers []Driver

	redisConnString := os.Getenv("REDISTOGO_URL")
	if len(redisConnString) == 0 {
		log.Printf(missingEnvVarErrString, "REDISTOGO_URL")
	} else {
		redisDriver, err := redis.New(redisConnString)
		if err == nil {
			drivers = append(drivers, redisDriver)
		}
	}

	// memcachedConnString := os.Getenv("MEMCACHED_CONN_STRING")
	// if len(memcachedConnString) == 0 {
	// 	log.Println("Missing MEMCACHED_CONN_STRING env var, are we?")
	// } else {
	// 	memcachedDriver, err := memcache.New(memcachedConnString)
	// 	if err == nil {
	// 		drivers = append(drivers, memcachedDriver)
	// 	}
	// }

	// MySQL
	// mySqlConnString := os.Getenv("MYSQL_CONN_STRING")
	// if len(mySqlConnString) == 0 {
	// 	// return nil, errors.New("Missing MYSQL_CONN_STRING env var, are we?")
	// }

	gostormInstance = *New(drivers...)

	log.Printf("Running server on %s", ServerAddr)
	http.Handle("/", configureRouter())
	err := http.ListenAndServe(ServerAddr, nil)
	if err != nil {
		panic(err)
	}
}
