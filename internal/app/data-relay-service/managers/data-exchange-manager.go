package managers

import (
	"bitbucket.org/project-alkemi/common-pkg/gen/consts"
	"bitbucket.org/project-alkemi/common-pkg/gen/messages"
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/tls-server"
	dr_discovery_service_api "bitbucket.org/project-alkemi/dr-discovery-service-api"
	"github.com/getsentry/sentry-go"
	"github.com/golang/protobuf/proto"
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

				var alkemiMessage messages.AlkemiMessage

				alkemiUnmarshalError := proto.UnmarshalText(msg.Payload, &alkemiMessage)

				if alkemiUnmarshalError != nil {

					if alkemiMessage.MessageType == consts.MessageType_CONTAINER_ORDER_FILLED {

						for _, conn := range dataExchangeManager.TLSServerMgr.AlkemiCompletedOrdersClients {

							go func(alkemiMessage *messages.AlkemiMessage, conn *tls_server.ClientConnections) {

								if conn.Active {

									streamSendError := conn.Stream.Send(alkemiMessage)

									logrus.Info("Sending message to: ", conn.Stream)

									if streamSendError != nil {
										logrus.Errorf("Error with Stream: %v - Error: %v", conn.Stream, streamSendError)
										conn.Active = false
										conn.Error <- streamSendError
									}

								}
							}(&alkemiMessage, conn)

						}
					}

					for _, conn := range dataExchangeManager.TLSServerMgr.AllDataClients {

						go func(alkemiMessage *messages.AlkemiMessage, conn *tls_server.ClientConnections) {

							if conn.Active {

								streamSendError := conn.Stream.Send(alkemiMessage)

								logrus.Info("Sending message to: ", conn.Stream)

								if streamSendError != nil {
									logrus.Errorf("Error with Stream: %v - Error: %v", conn.Stream, streamSendError)
									conn.Active = false
									conn.Error <- streamSendError
								}

							}
						}(&alkemiMessage, conn)

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
