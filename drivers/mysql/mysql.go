package mysql

import (
	"database/sql"
	"errors"
	"log"
	"os"

	// Well, go vet makes me comment on this.
	_ "github.com/go-sql-driver/mysql"
)

// Driver does the driving
type Driver struct {
	conn *sql.DB
}

// New Driver instance, right?
func New() (*Driver, error) {
	mySqlConnString := os.Getenv("MYSQL_CONN_STRING")
	if len(mySqlConnString) == 0 {
		return nil, errors.New("Missing MYSQL_CONN_STRING env var, are we?")
	}

	// Package mysql should:
	myConn, err := sql.Open("mysql", mySqlConnString)

	log.Printf("Connecting to MySQL => %s", mySqlConnString)

	if err != nil {
		return nil, errors.New("Can't connect to MySQL.")
	}

	log.Println("Connecting to MySQL => OK")

	return &Driver{conn: myConn}, nil
}

// Get the value!
func (drv *Driver) Get() {

}
