package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"database/sql"
	"database/sql/driver"

	_ "github.com/godror/godror"
	_ "github.com/lib/pq"
)

type DBInfo struct {
	DBType   string
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
	SelectListWithTailQuery(ctx Context, data DataBaseObject, tailQuery *TailQuery) (any, Error)
	ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error)
	SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error)

	CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error)
	CountRecordInTableWithTailQuery(ctx Context, data DataBaseObject, tailQuery *TailQuery) (int64, Error)
}

func openDBConnection(dbInfo DBInfo) dbSession {
	var session dbSession
	if dbInfo.DBType == DB_TYPE_POSTGRES {
		session = openPostgresDBConnection(dbInfo)
	} else if dbInfo.DBType == DB_TYPE_ORACLE {
		session = openOracleDBConnection(dbInfo)
	}
	return session
}

func openPostgresDBConnection(dbInfo DBInfo) *postgresSession {
	// Connect to postgres database and return session
	connectStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbInfo.Username, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database)

	fmt.Printf("Connect to postgres database: %s:%d/%s\n", dbInfo.Host, dbInfo.Port, dbInfo.Database)
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		log.Panicf("Connect to database fail: dbInfo = %v, err = %v", dbInfo, err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("Cannot ping to database: dbInfo = %v, err = %v", dbInfo, err)
	}

	fmt.Println("Connected to postgres database!")

	// Optionally, you can use an ORM like GORM to simplify the database operations
	return &postgresSession{
		DB: db,
	}
}

func connectToOracleDB(dbInfo DBInfo) (*sql.DB, Error) {
	connectStr := fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s"`, dbInfo.Username, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database)
	// Connect to oracle database and return session
	db, err := sql.Open("godror", connectStr)
	if err != nil {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, fmt.Sprintf("Error opening oracle database: dbInfo = %v, err = %v", dbInfo, err))
	}

	err = db.Ping()
	if err != nil {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, fmt.Sprintf("Error opening oracle database: dbInfo = %v, err = %v", dbInfo, err))
	}

	return db, nil
}

func openOracleDBConnection(dbInfo DBInfo) *oracleSession {
	db, err := connectToOracleDB(dbInfo)
	if err != nil {
		log.Panicf("Error opening oracle database: err = %v", err)
	}

	return &oracleSession{
		DB:         db,
		queryCount: 0,
		DBInfo:     dbInfo,
	}
}

func resetOracleSession(oracleSession *oracleSession) {
	LogInfo("Reset oracle session")
	err := oracleSession.Close()
	if err != nil {
		LogError("Error close oracle session when reset oracle session: %v", err)
	}
	oracleSession.resetOracleSession()
	newSesison := openOracleDBConnection(oracleSession.DBInfo)

	oracleSession.DB = newSesison.DB
	LogInfo("Reset oracle session success")
}
