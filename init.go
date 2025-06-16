package core

import (
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var mainDbSession dbSession
var secondaryDbSession dbSession
var commonApiMiddlewares []ApiMiddleware
var contextPool sync.Pool
var httpContextPool sync.Pool
var websocketContextPool sync.Pool
var Config CoreConfig
var redisClient cacheClient
var queueClient natsClient
var coreContext Context
var validate *validator.Validate
var contextTimeout time.Duration
var emqxBrokerClient MqttClient
var lockerManagerInstance *lockManager
var esClient searchClient

var router *gin.Engine

func Init(configFile string) {
	// Init core context
	coreContext = &rootContext{
		Context: context.Background(),
	}

	// Init config
	// Read config from file
	Config = loadConfigFile(configFile)
	if Config.Context.Timeout > 0 {
		contextTimeout = time.Second * time.Duration(Config.Context.Timeout)
	} else {
		// Default
		contextTimeout = time.Second * 60
	}

	// Set default if it is not config
	if Config.HttpClient.RetryTimes == 0 {
		Config.HttpClient.RetryTimes = 3
	}

	if Config.HttpClient.WaitTimes == 0 {
		Config.HttpClient.WaitTimes = 2000
	}

	// Init cache client
	if Config.Redis.Use {
		redisClient = connectCacheDB()
	}

	// Init database connection
	if Config.Database.Use {
		mainDbSession = openDBConnection(DBInfo{
			Host:     Config.Database.Host,
			Port:     int32(Config.Database.Port),
			Username: Config.Database.Username,
			Password: Config.Database.Password,
			Database: Config.Database.DatabaseName,
			DBType:   Config.Database.DBType,
		})
	}

	if Config.SecondaryDatabase.Use {
		secondaryDbSession = openDBConnection(DBInfo{
			Host:     Config.SecondaryDatabase.Host,
			Port:     int32(Config.SecondaryDatabase.Port),
			Username: Config.SecondaryDatabase.Username,
			Password: Config.SecondaryDatabase.Password,
			Database: Config.SecondaryDatabase.DatabaseName,
			DBType:   Config.SecondaryDatabase.DBType,
		})
	}

	// Init rabbitmq client
	if Config.NatsQueue.Use {
		queueClient = connectToNatsQueue(Config.NatsQueue.Url)
	}

	// Init emqx client
	if Config.Emqx.Use {
		emqxBrokerClient = NewEmqxClient(Config.Emqx)
	}

	// Init Elasticsearch client
	if Config.Elasticsearch.Use {
		esClient = connectElasticsearch()
	}

	// Init id generator
	initIdGenerator()
	// Core context will hold first id from instance
	coreContext.(*rootContext).contextID = ID.GenerateID()

	// Context pool
	contextPool = sync.Pool{
		New: func() any {
			return &rootContext{
				contextID:  BLANK,
				timeout:    0,
				cancelFunc: nil,
			}
		},
	}

	// Http context pool
	httpContextPool = sync.Pool{
		New: func() any {
			return &HttpContext{
				requestBody:    make([]byte, 16384),
				urlParams:      make(map[string]string),
				responseHeader: make(map[string][]string),
			}
		},
	}

	websocketContextPool = sync.Pool{
		New: func() any {
			return &websocketContext{
				requestID:  BLANK,
				timeout:    time.Duration(0),
				cancelFunc: nil,
				tempData:   make(map[string]any),
			}
		},
	}

	commonApiMiddlewares = make([]ApiMiddleware, 0)
	validate = validator.New()

	// Set background job
	interval := 30 * time.Second
	if Config.Scheduler.Interval != 0 {
		interval = time.Second * time.Duration(Config.Scheduler.Interval)
	}

	if Config.Scheduler.BucketSize == 0 {
		Config.Scheduler.BucketSize = 60
	}

	// Start Scheduler
	if Config.Scheduler.Use {
		startScheduler(interval)
	}

	callback = make(map[string]CallbackFunc)

	lockerManagerInstance = newLockManager()

	if Config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()

	if Config.HtmlFolder.Use {
		templ := template.New(BLANK).Funcs(basicFunctionMap)

		for _, pattern := range Config.HtmlFolder.Path {
			templ.ParseGlob(pattern)
		}

		router.SetHTMLTemplate(templ)
	}

	if Config.UseCorsOrigin {
		router.Use(corsMiddleware())
	}

	api := router.Group("/api")
	api.Use(TimeoutMiddleware(contextTimeout))
}

/*
* Release: Release all resources
* @return void
 */
func Release() {
	closeDB()
	releaseCacheDB()
	releaseMessageQueue()
	stopScheduler()
	if esClient.Client != nil {
		esClient.Close()
	}
}

func closeDB() {
	if mainDbSession != nil {
		mainDbSession.Close()
	}

	if secondaryDbSession != nil {
		secondaryDbSession.Close()
	}
}

func releaseCacheDB() {
	if redisClient.Client != nil {
		redisClient.Close()
	}
}

func releaseMessageQueue() {
	if queueClient.nc != nil {
		queueClient.nc.Close()
	}
}

/*
* Start: Start server
* Register all routes and listen to port
* @return void
 */
func Start() {
	// Callback function
	for _, cb := range callback {
		cb()
	}
	// Start secure server
	if Config.SecureServer.Use {
		go func() {
			err := router.RunTLS(":"+fmt.Sprintf("%d", Config.SecureServer.Port), Config.SecureServer.CertFile, Config.SecureServer.KeyFile)
			if err != nil {
				coreContext.LogFatal("Fail to start secure server. Error: %v", err)
			}
		}()
	}

	// Start http server
	router.Run(":" + fmt.Sprintf("%d", Config.Server.Port))
}

/*
* CacheClient: Get cache client
* @return cacheClient
 */
func CacheClient() cacheClient {
	return redisClient
}

/*
* MessageQueue: Get message queue client
* @return messageQueue
 */
func MessageQueue() natsClient {
	return queueClient
}

/*
* DBSession: Get database session
* @return dbSession
 */
func DBSession() dbSession {
	return mainDbSession
}

/*
* SecondaryDBSession: Get secondary database session
* @return dbSession
 */
func SecondaryDBSession() dbSession {
	return secondaryDbSession
}

func EmqxBrokerClient() MqttClient {
	return emqxBrokerClient
}

func ElasticsearchClient() searchClient {
	return esClient
}
