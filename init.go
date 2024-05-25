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

var pgSession dbSession
var routeMap map[string][]Route
var routeRegexMap map[string][]Route
var staticFolderMap map[string]staticFolder
var pageMap map[string]pageInfo
var httpContextPool sync.Pool
var contextPool sync.Pool
var commonApiMiddlewares []ApiMiddleware
var Config CoreConfig
var redisClient cacheClient
var rabbitMQClient *messageQueue
var coreContext Context
var validate *validator.Validate
var contextTimeout time.Duration
var htmlTemplateMap map[string]*template.Template

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
		pgSession = openDBConnection(DBInfo{
			Host:     Config.Database.Host,
			Port:     int32(Config.Database.Port),
			Username: Config.Database.Username,
			Password: Config.Database.Password,
			Database: Config.Database.DatabaseName,
		})
	}

	// Init rabbitmq client
	if Config.RabbitMQ.RetryTime == 0 {
		Config.RabbitMQ.RetryTime = 5
	}
	// Init rabbitmq client
	if Config.RabbitMQ.Use {
		rabbitMQClient = connectRabbitMQ()
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
	pgSession.Close()
}

func releaseCacheDB() {
	redisClient.Close()
}

func releaseMessageQueue() {
	rabbitMQClient.connection.Close()
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

	// Listen and serve
	LogInfo("Start server at port: %d", Config.Server.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Config.Server.Port), nil)
	if err != nil {
		log.Fatalln("ListenAndServe fail: ", err)
	}
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
func MessageQueue() *messageQueue {
	return rabbitMQClient
}

/*
* DBSession: Get database session
* @return dbSession
 */
func DBSession() dbSession {
	return pgSession
}

/*
* Handle API
 */

func handleAPIAndPage() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		isHandled := false
		if page, ok := pageMap[r.URL.Path]; ok && r.Method == http.MethodGet {
			LogInfo("Handle request page: %s", r.URL.Path)
			pageHandler(page, w, r)
			isHandled = true
		} else if routeList, ok := routeMap[r.URL.Path]; ok {
			LogInfo("Handle API: %s", r.URL.Path)
			for _, route := range routeList {
				if route.Method == r.Method {
					route.handler(w, r, optionalParams{})
					break
				}
			}
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
