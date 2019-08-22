package managers

import (
	dr_discovery_service_api "bitbucket.org/project-alkemi/dr-discovery-service-api"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/go-redis/redis"
)

type DataExchangeManager struct {
	RedisClient            *redis.Client
	CurrentSubscriptions   map[string]*redis.PubSub
	DiscoveryServiceApi    *dr_discovery_service_api.DrDiscoveryServiceApi
	LastServiceUpdateCheck time.Time
}

func CreateDataExchangeManager(redisClient *redis.Client, discoveryServiceApi *dr_discovery_service_api.DrDiscoveryServiceApi) *DataExchangeManager {

	return &DataExchangeManager{
		RedisClient:          redisClient,
		DiscoveryServiceApi:  discoveryServiceApi,
		CurrentSubscriptions: make(map[string]*redis.PubSub),
	}
}

func (dataExchangeManager *DataExchangeManager) ProcessServiceUpdates() {

	for {

		// Update the services list
		if time.Now().UTC().Sub(dataExchangeManager.LastServiceUpdateCheck).Seconds() > 5 {

			go dataExchangeManager.UpdateActiveServices()

			dataExchangeManager.LastServiceUpdateCheck = time.Now().UTC()
		}

		// Listen on data channels for response

		for key, value := range dataExchangeManager.CurrentSubscriptions {
			logrus.Info("Updating key: ", key)

			select {
			case msg := <-value.Channel():
				logrus.Info("Got : " + msg.Payload)
			default:
				logrus.Info("no activity")
			}

		}

		time.Sleep(time.Duration(100) * time.Millisecond)

	}
}

func (dataExchangeManager *DataExchangeManager) UpdateActiveServices() {

	allServicesResponse, allServicesResponseError := dataExchangeManager.DiscoveryServiceApi.RetrieveAllServices()

	if allServicesResponseError != nil {
		logrus.Error(allServicesResponseError)
		raven.CaptureError(allServicesResponseError, nil)
		return
	}

	// First we need to go through our current list and see if the channel is still in the services list

	for key, value := range dataExchangeManager.CurrentSubscriptions {

		if key != dataExchangeManager.DiscoveryServiceApi.UUID {

			found := false

			for _, element := range allServicesResponse.Services {
				if key == element.UUID {
					found = true
				}
			}

			if found == false {

				// Not found so we need to kill the connection and remove from current subscriptions
				pubSubCloseError := value.Close()
				if pubSubCloseError != nil {
					logrus.Error(pubSubCloseError)
					raven.CaptureError(pubSubCloseError, nil)
				}

				logrus.Info("UnSubscribing from: ", key)

				delete(dataExchangeManager.CurrentSubscriptions, key)
			}
		}
	}

	// Now we do the opposite go through the new services and if we are not subscribed we subscribe
	for _, element := range allServicesResponse.Services {

		if element.UUID != dataExchangeManager.DiscoveryServiceApi.UUID {

			found := false

			for key, _ := range dataExchangeManager.CurrentSubscriptions {

				if element.UUID == key {
					found = true
				}

			}

			if found == false {
				// We need to create a new subscription and add to the map
				pubsub := dataExchangeManager.RedisClient.Subscribe(element.UUID)

				_, subscriptionConfirmationError := pubsub.Receive()

				if subscriptionConfirmationError != nil {
					logrus.Error(subscriptionConfirmationError)
					raven.CaptureError(subscriptionConfirmationError, nil)
				}

				logrus.Info("Subscribing to: ", element.UUID, element.Alias)

				dataExchangeManager.CurrentSubscriptions[element.UUID] = pubsub
			}
		}
	}

}
