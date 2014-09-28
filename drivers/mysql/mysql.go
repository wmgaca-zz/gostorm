package mysql

import (
	"database/sql"
	"errors"
	"log"

	// Well, go vet makes me comment on this.
	_ "github.com/go-sql-driver/mysql"
)

// Driver does the driving
type Driver struct {
	conn *sql.DB
}

// New Driver instance, right?
func New(connString string) (*Driver, error) {
	// Package mysql should:
	myConn, err := sql.Open("mysql", connString)

	log.Printf("Connecting to MySQL => %s", connString)

	if err != nil {
		return nil, errors.New("Can't connect to MySQL.")
	}

	log.Println("Connecting to MySQL => OK")

	return &Driver{conn: myConn}, nil
}

// Get the value!
func (drv *Driver) Get() {

}
