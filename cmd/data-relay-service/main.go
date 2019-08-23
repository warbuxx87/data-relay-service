package main

import (
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/api"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/managers"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/models"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/util"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/views"
	dr_discovery_service_api "bitbucket.org/project-alkemi/dr-discovery-service-api"
	"encoding/gob"
	"github.com/getsentry/raven-go"
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

	config, configLoadError := util.LoadConfigurationFile()

	if configLoadError == nil {

		// Set the logrus config level
		if config.LogLevel == 0 {
			logrus.SetLevel(logrus.InfoLevel)
			logrus.Info("Log Level: Info")
		} else {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Info("Log Level: Debug")

		}

		databaseConnection, databaseError := connectToDatabase(config)

		if databaseError != nil {
			logrus.Error(databaseError, nil)
			return
		} else {
			logrus.Info("Successfully connected to database")
		}

		// Create the data context
		dataContext := util.DataContext{}

		dataContext.Config = config
		dataContext.Database = databaseConnection

		// Connect to sentry
		sentrySetDSNError := raven.SetDSN(config.SentryUrl)

		if sentrySetDSNError != nil {
			logrus.WithFields(logrus.Fields{}).Error(sentrySetDSNError)
		}

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
			return
		} else {
			logrus.Info("Successfully connected to redis server")
		}

		defer func() {
			redisConnCloseError := dataContext.RedisClient.Close()

			if redisConnCloseError != nil {
				logrus.Error(redisConnCloseError)
			}
		}()

		// Connect to discovery service
		dataContext.DiscoveryServiceAPI = dr_discovery_service_api.CreateDrDiscoveryServiceApi(config.DataRelayDiscoveryServiceUrl,
			config.DataRelayDiscoveryRegisterKey, config.DataRelayDiscoveryServiceAlias)

		registerResponse, registerResponseError := dataContext.DiscoveryServiceAPI.RegisterService()

		if registerResponseError != nil {
			logrus.Fatal(registerResponseError)
			return
		}

		if registerResponse.Success == false {
			logrus.Error("Failed to register to discovery service")
			logrus.Fatal(registerResponse.ErrorMessage)
			return
		} else {
			logrus.Info("Successfully registered with discovery service")
		}

		defer func() {
			unRegisterResponse, unRegisterResponseError := dataContext.DiscoveryServiceAPI.UnRegisterService()

			if unRegisterResponseError != nil {
				logrus.Error("Failed to unregister with discovery service")
				logrus.Error(unRegisterResponseError)
			}

			if unRegisterResponse.Success == false {
				logrus.Error("Failed to unregister with discovery service")
				logrus.Error(unRegisterResponse.ErrorMessage)
			} else {
				logrus.Info("UnRegistered with discovery service")
			}
		}()

		// Create the server manager

		dataContext.TLSServerManager = managers.CreateTLSServerManager(dataContext.DiscoveryServiceAPI.UUID, dataContext.RedisClient)

		dataContext.DataExchangeMgr = managers.CreateDataExchangeManager(dataContext.TLSServerManager, dataContext.RedisClient, dataContext.DiscoveryServiceAPI)

		go dataContext.DataExchangeMgr.ProcessServiceUpdates()

		// Start the checkin function. We checkin every 10 seconds
		go func() {

			for {

				checkInResponse, checkInResponseError := dataContext.DiscoveryServiceAPI.CheckIn()

				if checkInResponseError != nil {
					logrus.Error("CheckIn with discovery service failed")
					logrus.Error(checkInResponseError)
					return
				}

				if checkInResponse.Success == false {
					logrus.Error("CheckIn with discovery service failed")
					logrus.Error(checkInResponse.ErrorMessage)
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

		// Start tls server
		go func() {
			startTLSSeverError := dataContext.TLSServerManager.StartSever(dataContext.Config.TLSSeverIp,
				dataContext.Config.TLSSeverPort, dataContext.Config.TLSSeverCertLocation, dataContext.Config.TLSSeverKeyLocation)

			if startTLSSeverError != nil {
				logrus.Fatal(startTLSSeverError)
			}
		}()

		e.Logger.Fatal(e.Start(config.ListenIp + ":" + strconv.Itoa(config.ListenPort)))

	}
}
