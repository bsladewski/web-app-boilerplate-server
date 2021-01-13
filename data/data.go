// Package data exposes a database handle for persistent storage.
//
// Environment:
//     WEB_APP_CONNECTION_STRING
//         string - the connection string to establish a database connection
//     WEB_APP_IN_MEMORY_DATABASE
//         boolean - use an in-memory database in place of a real database
//     WEB_APP_USE_MOCK_DATA:
//         boolean - populate database with mock data on startup
package data

import (
	"os"
	"strings"

	"web-app/env"

	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// init establishes a connection to the application database or sets up an
// in-memory database for testing.
func init() {

	// only initialize the database if it has not already been initialized
	if db != nil {
		return
	}

	// check if we are running a unit test
	test := strings.HasSuffix(os.Args[0], ".test")

	// check if we should use an in-memory database
	inMemory := env.GetBoolSafe(inMemoryDatabaseVariable, false)

	var err error

	if test || inMemory {
		// if this is a test or the in-memory environment variable is set create
		// an in-memory application database
		db, err = gorm.Open(
			sqlite.Open(inMemoryConnectionString),
			&gorm.Config{},
		)
	} else {
		// establish a connection to the application database
		db, err = gorm.Open(
			mysql.Open(env.MustGetString(connectionStringVariable)),
			&gorm.Config{},
		)
	}

	if err != nil {
		// if we were unable to establish a database connection log the error
		// and exit the application
		logrus.Fatal(err)
	}

	// check if we should load mock data
	useMockData = env.GetBoolSafe(useMockDataVariable, false)

}

const (
	// connectionStringVariable defines an environment variable for the MySQL
	// connection string used to establish a connection to the application
	// database.
	connectionStringVariable = "WEB_APP_CONNECTION_STRING"
	// inMemoryDatabaseVariable defines an environment variable that, if set to
	// true, will replace the database connection with an in-memory database.
	inMemoryDatabaseVariable = "WEB_APP_IN_MEMORY_DATABASE"
	// useMockDataVariable defines an environment variable that, if set to true,
	// will load mock data.
	useMockDataVariable = "WEB_APP_USE_MOCK_DATA"
	// inMemoryConnectionString defines the string that will be used to create
	// an in-memory database for testing.
	inMemoryConnectionString = "file::memory:?cache=shared"
)

// db is used to work with persistent storage.
var db *gorm.DB

// useMockData is used to determine if mock data should be loaded.
var useMockData bool

// DB retrieves a handle to the application database.
func DB() *gorm.DB {
	return db
}

// Ping performs a simple against the database to check availability.
func Ping() error {
	return db.Raw(`SELECT 1;`).Error
}

// UseMockData checks whether the current runtime should load mock data.
func UseMockData() bool {
	return useMockData
}
