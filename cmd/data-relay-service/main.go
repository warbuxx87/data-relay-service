package main

import (
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/api"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/managers"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/models"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/tls-server"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/util"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/views"
	dr_discovery_service_api "bitbucket.org/project-alkemi/dr-discovery-service-api"
	"encoding/gob"
	"flag"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"html/template"
	"os"
	"strconv"
	"time"
)

func connectToDatabase(config *util.Configuration) (*gorm.DB, error) {

	database, databaseError := gorm.Open("sqlite3", "datacache.db")

	if databaseError != nil {
		return nil, databaseError
	}

	database.AutoMigrate(&models.UserModel{})

	// Check to see if the default user exit
	util.CheckForDefaultUser(database, config)

	return database, nil

}

func init() {

	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.InfoLevel)
}

func main() {

	util.ShowBanner()

	prod := flag.Bool("prod", false, "Use '$go run ./cmd/... -prod' to set to true, changing default configuration file from 'dev.json' to 'config.json'")
	flag.Parse()

	var configFilename string
	if *prod {
		configFilename = "config.json"
	} else {
		configFilename = "dev.json"
	}

	config, configLoadError := util.LoadConfigurationFile(configFilename)
	if configLoadError != nil {
		logrus.Fatalf("failed to load app configuration; %v", configLoadError)
		os.Exit(1)
	}

	// Set the logrus config level
	if config.LogLevel == 0 {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("Log Level: Info")
	} else {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("Log Level: Debug")
	}

	// Connect to sentry
	sentrySetDSNError := sentry.Init(sentry.ClientOptions{
		Dsn: config.SentryUrl,
	})

	if sentrySetDSNError != nil {
		logrus.WithFields(logrus.Fields{}).Error(sentrySetDSNError)
	}

	databaseConnection, databaseError := connectToDatabase(config)
	if databaseError != nil {
		logrus.Error(databaseError, nil)
		sentry.CaptureException(databaseError)
		return
	} else {
		logrus.Info("Successfully connected to database")
	}

	// Create the data context
	dataContext := util.DataContext{}

	dataContext.Config = config
	dataContext.Database = databaseConnection

	// Create the redis connection
	dataContext.RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisIP + ":" + strconv.Itoa(config.RedisPort),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, redisConnectionError := dataContext.RedisClient.Ping().Result()

	if redisConnectionError != nil {
		logrus.Info("Failed to connect to redis server")

		logrus.Fatal(redisConnectionError)
		sentry.CaptureException(redisConnectionError)
		return
	} else {
		logrus.Info("Successfully connected to redis server")
	}

	defer func() {
		redisConnCloseError := dataContext.RedisClient.Close()

		if redisConnCloseError != nil {
			logrus.Error(redisConnCloseError)
			sentry.CaptureException(redisConnCloseError)
		}
	}()

	// Connect to discovery service
	dataContext.DiscoveryServiceAPI = dr_discovery_service_api.CreateDrDiscoveryServiceApi(config.DataRelayDiscoveryServiceUrl,
		config.DataRelayDiscoveryRegisterKey, config.DataRelayDiscoveryServiceAlias)

	registerResponse, registerResponseError := dataContext.DiscoveryServiceAPI.RegisterService()

	if registerResponseError != nil {
		logrus.Fatal(registerResponseError)
		sentry.CaptureException(registerResponseError)
		return
	}

	if registerResponse.Success == false {
		logrus.Error("Failed to register to discovery service")
		logrus.Fatal(registerResponse.ErrorMessage)
		sentry.CaptureMessage(registerResponse.ErrorMessage)
		return
	} else {
		logrus.Info("Successfully registered with discovery service")
	}

	defer func() {
		unRegisterResponse, unRegisterResponseError := dataContext.DiscoveryServiceAPI.UnRegisterService()

		if unRegisterResponseError != nil {
			logrus.Error("Failed to unregister with discovery service")
			logrus.Error(unRegisterResponseError)
			sentry.CaptureException(unRegisterResponseError)
		}

		if unRegisterResponse.Success == false {
			logrus.Error("Failed to unregister with discovery service")
			logrus.Error(unRegisterResponse.ErrorMessage)
			sentry.CaptureMessage(unRegisterResponse.ErrorMessage)
		} else {
			logrus.Info("UnRegistered with discovery service")
		}
	}()

	// Create the server manager

	dataContext.TLSServerManager = tls_server.CreateTLSServerManager(dataContext.DiscoveryServiceAPI.UUID, dataContext.Config.DataRelayDiscoveryServiceAlias, dataContext.RedisClient)

	dataContext.DataExchangeMgr = managers.CreateDataExchangeManager(dataContext.TLSServerManager, dataContext.RedisClient, dataContext.DiscoveryServiceAPI)

	go dataContext.DataExchangeMgr.ProcessServiceUpdates()
	go dataContext.TLSServerManager.CheckRedisConnection()

	go api.ContinuousHealthCheck(&dataContext)

	// Start the checkin function. We checkin every 10 seconds
	go func() {

		for {

			checkInResponse, checkInResponseError := dataContext.DiscoveryServiceAPI.CheckIn()

			if checkInResponseError != nil {
				logrus.Error("CheckIn with discovery service failed")
				logrus.Error(checkInResponseError)
				sentry.CaptureException(checkInResponseError)
				return
			}

			if checkInResponse.Success == false {
				logrus.Error("CheckIn with discovery service failed")
				logrus.Error(checkInResponse.ErrorMessage)
				sentry.CaptureMessage(checkInResponse.ErrorMessage)
			}

			time.Sleep(time.Duration(10) * time.Second)
		}
	}()

	e := echo.New()
	e.Debug = true

	// Register our custom app context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &util.EndpointContext{Context: c, AppContext: &dataContext}
			return h(cc)
		}
	})

	e.Static("/static", "assets")

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(uuid.NewV4().String()))))
	e.Use(middleware.Logger())

	templates := make(map[string]*template.Template)

	templates["login.html"] = template.Must(template.ParseFiles("templates/login.html", "templates/base.html"))
	templates["dashboard.html"] = template.Must(template.ParseFiles("templates/dashboard.html", "templates/base.html"))

	e.Renderer = &util.TemplateRegistry{
		Templates: templates,
	}

	gob.Register(models.UserModel{})

	e.HideBanner = true

	e.GET("/", views.RedirectView)
	e.GET("/login", views.LoginView)
	e.GET("/dashboard", views.DashboardView)

	e.GET("/action/logout", api.LogoutAction)
	e.POST("/action/login", api.LoginAction)
	e.GET("/health", api.HealthCheckAction)

	// Start tls server
	go func() {
		startTLSServerError := dataContext.TLSServerManager.StartSever(dataContext.Config.TLSSeverIp,
			dataContext.Config.TLSSeverPort, dataContext.Config.TLSSeverCertLocation, dataContext.Config.TLSSeverKeyLocation)

		if startTLSServerError != nil {
			logrus.Fatal(startTLSServerError)
			sentry.CaptureException(startTLSServerError)
		}
	}()

	e.Logger.Fatal(e.Start(config.ListenIp + ":" + strconv.Itoa(config.ListenPort)))
}
