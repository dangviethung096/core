package core

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DBInfo struct {
	DBType   string
	Host     string
	Port     int32
	Username string
	Password string
	Database string
	Timeout  int64
	SSLMode  string
}

type DataBaseObject interface {
	TableName() string
}

type dbSession interface {
	// Command
	InsertDataToDB(ctx Context, data DataBaseObject) Error
	DeleteDataFromDBByID(ctx Context, data DataBaseObject) Error
	DeleteDataFromDBWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error
	UpdateDataToDB(ctx Context, data DataBaseObject) Error

	// Query
	SelectListByFields(ctx Context, data DataBaseObject, whereQuery string, args ...any) (any, Error)
	SelectListByFieldWithPaging(ctx Context, data DataBaseObject, limit int64, offset int64, whereQuery string, args ...any) (any, Error)
	SelectPaging(ctx Context, data DataBaseObject, orderQuery string, limit int64, offset int64) (any, Error)
	SelectByID(ctx Context, data DataBaseObject) Error

	CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error)
	CountRecordInTableWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) (int64, Error)

	// onnection
	Connection() *gorm.DB
	GetOriginConnection(ctx Context) *sql.DB

	// Auto migrate
	AutoMigrate(ctx Context, data DataBaseObject) Error

	// Close
	Close()
}

func openDBConnection(dbInfo DBInfo) dbSession {
	var session dbSession
	if dbInfo.DBType == DB_TYPE_POSTGRES {
		session = openPostgresDBConnection(dbInfo)
	}
	return session
}

func openPostgresDBConnection(dbInfo DBInfo) *postgresSession {
	// Connect to postgres database and return session
	connectionStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", dbInfo.Host, dbInfo.Username, dbInfo.Password, dbInfo.Database, dbInfo.Port, dbInfo.SSLMode)

	log.Printf("Connect to postgres database: %s:%d/%s\n", dbInfo.Host, dbInfo.Port, dbInfo.Database)

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: gormlogger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
			gormlogger.Config{
				SlowThreshold:             time.Second,     // Slow SQL threshold
				LogLevel:                  gormlogger.Info, // Log level (Silent, Error, Warn, Info)
				IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,            // Enable color
				ParameterizedQueries:      false,           // Include params in the SQL log
			},
		),
	}

	db, err := gorm.Open(postgres.Open(connectionStr), gormConfig)
	if err != nil {
		log.Panicf("Cannot connect to database: dbInfo = %v, err = %v", dbInfo, err)
	}

	log.Printf("Connected to postgres database!\n")

	return &postgresSession{
		DB: db,
	}
}
