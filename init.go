package core

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/go-playground/validator"
)

var mainDbSession dbSession
var secondaryDbSession dbSession

var routeMap map[string][]Route
var routeRegexMap map[string][]Route
var uploadFileHandlerMap map[string]UploadFileHandler

var staticFolderMap map[string]staticFolder

var pageMap map[string]pageInfo

var commonApiMiddlewares []ApiMiddleware

var contextPool sync.Pool

var httpContextPool sync.Pool

var websocketContextPool sync.Pool
var websocketRouteMap map[string]websocketRoute

var Config CoreConfig
var redisClient cacheClient
var queueClient natsClient
var coreContext Context
var validate *validator.Validate
var contextTimeout time.Duration
var htmlTemplateMap map[string]*template.Template
var emqxBrokerClient MqttClient
var lockerManagerInstance *lockManager

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

	// Init id generator
	initIdGenerator()
	// Core context will hold first id from instance
	coreContext.(*rootContext).contextID = ID.GenerateID()

	routeMap = make(map[string][]Route)
	routeRegexMap = make(map[string][]Route)
	staticFolderMap = make(map[string]staticFolder)
	pageMap = make(map[string]pageInfo)
	htmlTemplateMap = make(map[string]*template.Template)
	uploadFileHandlerMap = make(map[string]UploadFileHandler)
	websocketRouteMap = make(map[string]websocketRoute)

	// Context pool
	contextPool = sync.Pool{
		New: func() interface{} {
			return &rootContext{
				contextID:  BLANK,
				timeout:    0,
				cancelFunc: nil,
			}
		},
	}

	// Http context pool
	httpContextPool = sync.Pool{
		New: func() interface{} {
			return &HttpContext{
				requestBody:    make([]byte, 16384),
				urlParams:      make(map[string]string),
				responseHeader: make(map[string][]string),
			}
		},
	}

	websocketContextPool = sync.Pool{
		New: func() interface{} {
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
	// Register all static folders
	handleStaticFolder()

	// Register all routes
	handleAPIAndPage()

	blockServerChan := make(chan string)
	// Listen and serve
	go func() {
		LogInfo("Start server at port: %d", Config.Server.Port)
		blockServerChan <- "Start server"
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", Config.Server.Port), nil)
		if err != nil {
			log.Fatalln("ListenAndServe fail: ", err)
		}
	}()

	go func() {
		if Config.SecureServer.Use {
			LogInfo("Start secure server at port: %d", Config.SecureServer.Port)
			err := http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", Config.SecureServer.Port), Config.SecureServer.CertFile, Config.SecureServer.KeyFile, nil)
			if err != nil {
				log.Fatalln("ListenAndServeTLS fail: ", err)
			}
		}
	}()

	<-blockServerChan
	// Callback function
	for _, cb := range callback {
		cb()
	}

	// Wait for stop server signal
	<-blockServerChan
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

/*
* Handle API
 */

func handleAPIAndPage() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		isHandled := false
		if page, ok := pageMap[r.URL.Path]; ok && r.Method == http.MethodGet {
			pageHandler(page, w, r)
			isHandled = true
		} else if routeList, ok := routeMap[r.URL.Path]; ok {
			LogInfo("Handle API: %s", r.URL.Path)
			for _, route := range routeList {
				if route.Method == r.Method {
					route.handler(w, r, optionalParams{})
					isHandled = true
					break
				}
			}
		} else if route, ok := websocketRouteMap[r.URL.Path]; ok {
			route.handler(w, r)
			isHandled = true
		} else if handler, ok := uploadFileHandlerMap[r.URL.Path]; ok {
			handler.handler(w, r)
			isHandled = true
		} else {
			for regexPath, routeList := range routeRegexMap {
				if match, _ := regexp.MatchString(regexPath, r.URL.Path); match {
					LogInfo("Handle Regex API: %s", r.URL.Path)
					for _, route := range routeList {
						if route.Method == r.Method {
							route.handler(w, r, optionalParams{
								haveUrlParam: true,
								urlPattern:   regexPath,
								urlParamKeys: route.URL.Params,
							})
							isHandled = true
							break
						}
					}

					if isHandled {
						break
					}
				}
			}
		}

		if !isHandled {
			http.NotFound(w, r)
		}

	})
}

/*
* Handle Static Folder
 */

func handleStaticFolder() {
	for _, staticFolder := range staticFolderMap {
		LogInfo("Register static folder: url = %s, path = %s", staticFolder.url, staticFolder.path)
		http.Handle(staticFolder.url, http.StripPrefix(staticFolder.prefix, http.FileServer(http.Dir(staticFolder.path))))
	}
}
