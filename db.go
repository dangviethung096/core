package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"database/sql"
	"database/sql/driver"

	_ "github.com/lib/pq"
)

type DBInfo struct {
	Host     string
	Port     int32
	Username string
	Password string
	Database string
	Timeout  int64
	// TODO
}

type dbSession interface {
	// Method for sql.DB
	PingContext(ctx context.Context) error
	Ping() error
	Close() error
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
	Stats() sql.DBStats
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Exec(query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRow(query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	Driver() driver.Driver
	Conn(ctx context.Context) (*sql.Conn, error)

	// Additional methods for dbSession
	SaveDataToDB(ctx Context, data DataBaseObject) Error
	SaveDataToDBWithoutPrimaryKey(ctx Context, data DataBaseObject) Error
	DeleteDataInDB(ctx Context, data DataBaseObject) Error
	// TODO: Should change in the future
	DeleteDataWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) Error
	UpdateDataInDB(ctx Context, data DataBaseObject) Error

	SelectById(ctx Context, data DataBaseObject) Error
	ListAllInTable(ctx Context, data DataBaseObject) (any, Error)
	SelectListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}) (any, Error)
	SelectListWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) (any, Error)
	ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error)
	SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error)

	CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error)
	CountRecordInTableWithWhere(ctx Context, data DataBaseObject, whereQuery string) (int64, Error)
}

func (info *DBInfo) buildConnectionString() string {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", info.Username, info.Password, info.Host, info.Port, info.Database)
	// Configure the database connection string with the host, port, user, password, and dbname details
	return connStr
}

func openDBConnection(dbInfo DBInfo) dbSession {
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

	fmt.Println("Connected to database!")

	// Optionally, you can use an ORM like GORM to simplify the database operations
	return postgresSession{db}
}
