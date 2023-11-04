package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/wgroup1/enigma/middleware"
)

var mode string = "on-premmise"

func NewMYSQLConn() *sql.DB{
	
	if os.Getenv("ENV_MODE") != "" {
		mode = os.Getenv("ENV_MODE")
	}

	switch mode {
	case "cloud":
		dBConnection, err := connectUnixSocket()

		if err != nil {
			log.Println("Connection Failed!!")
		}
		err = dBConnection.Ping()
		if err != nil {
			log.Printf("Ping Failed!! %v", err)
		}

		dBConnection.SetMaxOpenConns(10000)
		dBConnection.SetMaxIdleConns(0)
		dBConnection.SetConnMaxLifetime(time.Minute * 4)

		return dBConnection
	default:
		dBConnection, err := sql.Open("mysql", middleware.GetViperEnvVariable("DB_URL"))
		// dBConnection, err := sql.Open("mysql", os.Getenv("DB_URL"))

		if err != nil {
			log.Println("Connection Failed!!")
		}
		err = dBConnection.Ping()
		if err != nil {
			log.Printf("Ping Failed!! %v", err)
		}


		dBConnection.SetMaxIdleConns(50)
		dBConnection.SetMaxOpenConns(150)
		dBConnection.SetConnMaxLifetime(120 * time.Minute)
		dBConnection.SetConnMaxIdleTime(5 * time.Minute)
		return dBConnection

	}
}

func GetViperEnvVariable(s string) {
	panic("unimplemented")
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