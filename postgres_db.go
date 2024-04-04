package core

import (
	"fmt"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

type postgresDBInfo struct {
	Host     string
	Port     int32
	Username string
	Password string
	Database string
	Timeout  int64
	// TODO
}

type postgresDBSession struct {
	*sql.DB
}

func (info *postgresDBInfo) buildConnectionString() string {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", info.Username, info.Password, info.Host, info.Port, info.Database)
	// Configure the database connection string with the host, port, user, password, and dbname details
	return connStr
}

func openPostgresDBConnection(dbInfo postgresDBInfo) postgresDBSession {
	// Connect to postgres database and return session
	connectStr := dbInfo.buildConnectionString()
	fmt.Printf("Connect to postgres database: %s:%d/%s\n", dbInfo.Host, dbInfo.Port, dbInfo.Database)
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		log.Panicf("Connect to database fail: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("Cannot ping to database: %v", err)
	}

	// Optionally, you can use an ORM like GORM to simplify the database operations
	return postgresDBSession{db}
}
