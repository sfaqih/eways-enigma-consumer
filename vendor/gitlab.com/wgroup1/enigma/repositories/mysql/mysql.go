package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"

	//get mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var DBUrl string = ""
var STATIC_TOKEN string = ""

// ViperEnvVariable is func to get .env file
func ViperEnvVariable(key string) string {
	switch key {
	case "DB_URL":
		if DBUrl != "" {
			return DBUrl
		}
	case "STATIC_TOKEN":
		if STATIC_TOKEN != "" {
			return STATIC_TOKEN
		}
	}

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("Error while reading config file:", err)
	}

	value, ok := viper.Get(key).(string)

	if !ok {
		fmt.Println("Invalid type assertion")
	}

	switch key {
	case "DB_URL":
		DBUrl = value
	case "STATIC_TOKEN":
		STATIC_TOKEN = value
	}

	return value
}

// InitializeMySQL is the func to open the connection for MySQL
func InitializeMySQL() *sql.DB {
	var mode string = "on-premmise"
	if os.Getenv("ENV_MODE") != "" {
		mode = os.Getenv("ENV_MODE")
	}

	switch mode {
	case "cloud":
		dBConnection, err := connectUnixSocket()

		if err != nil {
			fmt.Println("Connection Failed!!")
		}
		err = dBConnection.Ping()
		if err != nil {
			fmt.Printf("Ping Failed!! %v", err)
		}

		dBConnection.SetMaxOpenConns(10000)
		dBConnection.SetMaxIdleConns(0)
		dBConnection.SetConnMaxLifetime(time.Minute * 4)

		return dBConnection
	default:
		dBConnection, err := sql.Open("mysql", ViperEnvVariable("DB_URL"))
		// dBConnection, err := sql.Open("mysql", os.Getenv("DB_URL"))

		if err != nil {
			fmt.Println("Connection Failed!!")
		}
		err = dBConnection.Ping()
		if err != nil {
			fmt.Printf("Ping Failed!! %v", err)
		}

		dBConnection.SetMaxOpenConns(10000)
		dBConnection.SetMaxIdleConns(0)
		dBConnection.SetConnMaxLifetime(time.Minute * 4)
		return dBConnection

	}

}

func connectUnixSocket() (*sql.DB, error) {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Warning: %s environment variable not set.", k)
		}
		return v
	}
	// Note: Saving credentials in environment variables is convenient, but not
	// secure - consider a more secure solution such as
	// Cloud Secret Manager (https://cloud.google.com/secret-manager) to help
	// keep secrets safe.
	var (
		dbUser         = mustGetenv("DB_USER")              // e.g. 'my-db-user'
		dbPwd          = mustGetenv("DB_PASS")              // e.g. 'my-db-password'
		dbName         = mustGetenv("DB_NAME")              // e.g. 'my-database'
		unixSocketPath = mustGetenv("INSTANCE_UNIX_SOCKET") // e.g. '/cloudsql/project:region:instance'
	)

	dbURI := fmt.Sprintf("%s:%s@unix(/%s)/%s?parseTime=true",
		dbUser, dbPwd, unixSocketPath, dbName)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// ...

	return dbPool, nil
}

// CloseRows is func to close the rows result from SELECT query
func CloseRows(rows *sql.Rows) {
	if rows != nil {
		rows.Close()
	}
}

// CloseStmt is func to prepare the statement
func CloseStmt(stmt *sql.Stmt) {
	if stmt != nil {
		stmt.Close()
	}
}
