package managers

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	common_message_models "bitbucket.org/project-alkemi/common-message-pkg/common-message-models"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/tls-server"
	dr_discovery_service_api "bitbucket.org/project-alkemi/dr-discovery-service-api"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/go-redis/redis"
)

type DataExchangeManager struct {
	RedisClient            *redis.Client
	CurrentSubscriptions   map[string]*redis.PubSub
	DiscoveryServiceApi    *dr_discovery_service_api.DrDiscoveryServiceApi
	LastServiceUpdateCheck time.Time
	TLSServerMgr           *tls_server.TLSServerManager
}

func CreateDataExchangeManager(tlsServerManager *tls_server.TLSServerManager, redisClient *redis.Client, discoveryServiceApi *dr_discovery_service_api.DrDiscoveryServiceApi) *DataExchangeManager {

	return &DataExchangeManager{
		RedisClient:          redisClient,
		DiscoveryServiceApi:  discoveryServiceApi,
		TLSServerMgr:         tlsServerManager,
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

			select {
			case msg := <-value.Channel():

				logrus.Info("Got payload from : " + key)

				decodedBytes, decodedBytesError := base64.StdEncoding.DecodeString(msg.Payload)

				if decodedBytesError != nil {
					logrus.Error(decodedBytesError)
					sentry.CaptureException(decodedBytesError)
				}

				var rawMessage common_message_models.RawMessageModel

				orderFilledMessageMarshalError := gob.NewDecoder(bytes.NewBuffer(decodedBytes)).Decode(&rawMessage)

				if orderFilledMessageMarshalError != nil {
					logrus.Error(orderFilledMessageMarshalError)
					sentry.CaptureException(orderFilledMessageMarshalError)
				}

				rawMessage.FromDataRelay = true
				rawMessage.FromDataRelayAlias = dataExchangeManager.DiscoveryServiceApi.Alias

				// Broadcast to all clients
				for key, _ := range *dataExchangeManager.TLSServerMgr.ConnectedClients {

					if dataExchangeManager.TLSServerMgr.DataChannels.CheckIfChannelSubscribed((*key).RemoteAddr().String(), common_message_consts.DataChannelType_AllData) {

						clientEnc := gob.NewEncoder(*key)

						writeDataToSocketError := clientEnc.Encode(rawMessage)

						if writeDataToSocketError != nil {
							logrus.Error(writeDataToSocketError)
							sentry.CaptureException(writeDataToSocketError)
						}
					}

				}
			default:
				break
			}

		}

		time.Sleep(time.Duration(1) * time.Millisecond)

	}
}

func (dataExchangeManager *DataExchangeManager) UpdateActiveServices() {

	allServicesResponse, allServicesResponseError := dataExchangeManager.DiscoveryServiceApi.RetrieveAllServices()

	if allServicesResponseError != nil {
		logrus.Error(allServicesResponseError)
		sentry.CaptureException(allServicesResponseError)
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
					sentry.CaptureException(pubSubCloseError)
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
					sentry.CaptureException(subscriptionConfirmationError)
				}

				logrus.Info("Subscribing to: ", element.UUID, element.Alias)

				dataExchangeManager.CurrentSubscriptions[element.UUID] = pubsub
			}
		}
	}

}
