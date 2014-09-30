package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Debug mode, more verbose if true
var Debug = true

var gostormInstance Gostorm

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
			log.Printf("gs.ret => %s", ret)
			return ret, nil
		case err = <-errChan:
			retCount++
			log.Printf("gs.err => %s", err)
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

func configureRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/get/{key:[a-zA-Z0-9:.]+}/", getHandler).Methods("GET")

	return router
}

func main() {
	// 	log.Println("Starting gostorm...")
	//
	// 	var drivers []Driver
	//
	// 	redisConnString := os.Getenv("REDIS_CONN_STRING")
	// 	if len(redisConnString) == 0 {
	// 		log.Println("Missing REDIS_CONN_STRING env var, are we?")
	// 	} else {
	// 		redisDriver, err := redis.New(redisConnString)
	// 		if err == nil {
	// 			drivers = append(drivers, redisDriver)
	// 		}
	// 	}
	//
	// 	memcachedConnString := os.Getenv("MEMCACHED_CONN_STRING")
	// 	if len(memcachedConnString) == 0 {
	// 		log.Println("Missing MEMCACHED_CONN_STRING env var, are we?")
	// 	} else {
	// 		memcachedDriver, err := memcache.New(memcachedConnString)
	// 		if err == nil {
	// 			drivers = append(drivers, memcachedDriver)
	// 		}
	// 	}
	//
	// 	gostormInstance = *New(drivers...)
	//
	// 	// MySQL
	// 	mySqlConnString := os.Getenv("MYSQL_CONN_STRING")
	// 	if len(mySqlConnString) == 0 {
	// 		// return nil, errors.New("Missing MYSQL_CONN_STRING env var, are we?")
	// 	}

	ServerAddr := ":" + os.Getenv("PORT")
	log.Printf("Running server on %s", ServerAddr)

	http.Handle("/", configureRouter())
	// http.HandleFunc("/", homeHandler)
	fmt.Println("listening...")

	err := http.ListenAndServe(ServerAddr, nil)
	if err != nil {
		panic(err)
	}
}
