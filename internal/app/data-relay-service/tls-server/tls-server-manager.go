package tls_server

import (
	"bitbucket.org/project-alkemi/common-pkg/gen/messages"
	"bitbucket.org/project-alkemi/common-pkg/gen/services"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"strconv"
	"sync"
	"time"
)

type AllDataConnection struct {
	Stream services.DataRelayService_ReceiveAllDataServer
	UUID   string
	Active bool
	Error  chan error
}

type TLSServerManager struct {
	AllDataClients []*AllDataConnection
	RedisClient    *redis.Client
	ServiceUUID    string
	ServiceAlias   string
}

func CreateTLSServerManager(serviceUUID string, serviceAlias string, redisClient *redis.Client) *TLSServerManager {

	tlsServerManager := TLSServerManager{}

	var allDataClients []*AllDataConnection
	tlsServerManager.AllDataClients = allDataClients
	tlsServerManager.RedisClient = redisClient
	tlsServerManager.ServiceUUID = serviceUUID
	tlsServerManager.ServiceAlias = serviceAlias
	return &tlsServerManager
}

func (tlsServerManager *TLSServerManager) StartSever(tlsServerIP string, tlsServerPort int, tlsCertLocation string, tlsKeyLocation string) error {

	tlsCertConfig, loadKeyPairError := credentials.NewServerTLSFromFile(tlsCertLocation, tlsKeyLocation)

	if loadKeyPairError != nil {
		logrus.Error(loadKeyPairError)
		sentry.CaptureException(loadKeyPairError)
		return loadKeyPairError
	}

	gRPCServer := grpc.NewServer(grpc.Creds(tlsCertConfig))

	tlsServerListener, tlsServerListenerError := net.Listen("tcp", tlsServerIP+":"+strconv.Itoa(tlsServerPort))

	if tlsServerListenerError != nil {
		logrus.Info(tlsServerListenerError)
		return tlsServerListenerError
	}

	services.RegisterDataRelayServiceServer(gRPCServer, tlsServerManager)

	fmt.Println("⇨ tls sever started on " + tlsServerIP + ":" + strconv.Itoa(tlsServerPort))

	serveError := gRPCServer.Serve(tlsServerListener)

	if serveError != nil {
		logrus.Info(serveError)
		return serveError
	}

	return nil
}

func (tlsServerManager *TLSServerManager) ReceiveAllData(req *messages.Empty, stream services.DataRelayService_ReceiveAllDataServer) error {

	allDataConnection := &AllDataConnection{
		Stream: stream,
		UUID:   uuid.NewV4().String(),
		Active: true,
		Error:  make(chan error),
	}

	tlsServerManager.AllDataClients = append(tlsServerManager.AllDataClients, allDataConnection)

	logrus.Info("Received new connection request from: ", stream)

	return <-allDataConnection.Error
}

func (tlsServerManager *TLSServerManager) SendData(ctx context.Context, alkemiMessage *messages.AlkemiMessage) (*messages.RequestResponse, error) {

	wait := sync.WaitGroup{}
	done := make(chan int)

	logrus.Info("Received message from: ", alkemiMessage.AccessUUID)

	for _, conn := range tlsServerManager.AllDataClients {
		wait.Add(1)

		go func(alkemiMessage *messages.AlkemiMessage, conn *AllDataConnection) {
			defer wait.Done()

			if conn.Active {

				alkemiMessage.FromDataRelay = true
				alkemiMessage.FromDataRelayAlias = tlsServerManager.ServiceAlias

				streamSendError := conn.Stream.Send(alkemiMessage)

				logrus.Info("Sending message to: ", conn.Stream)

				if streamSendError != nil {
					logrus.Errorf("Error with Stream: %v - Error: %v", conn.Stream, streamSendError)
					conn.Active = false
					conn.Error <- streamSendError
				}

				redisClient := tlsServerManager.RedisClient
				if redisClient != nil {
					publishError := redisClient.Publish(tlsServerManager.ServiceUUID, proto.MarshalTextString(alkemiMessage)).Err()
					if publishError != nil {
						logrus.Error(publishError)
						sentry.CaptureException(publishError)
					}
				}

			}
		}(alkemiMessage, conn)

	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &messages.RequestResponse{Success: true, ErrorStr: ""}, nil

}

// CheckRedisConnection pings redis server every 10secs.
// it'll report error to sentry if there was any
func (tlsServerManager *TLSServerManager) CheckRedisConnection() {
	interval := 10 * time.Second
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			logrus.Info("pinging redis server...")
			result, redisError := tlsServerManager.RedisClient.Ping().Result()
			if redisError != nil {
				logrus.Warn("redis PING command failed: ", redisError)
				sentry.CaptureException(errors.Wrap(redisError, "redis PING command failed"))
			} else {
				logrus.Info("Redis server response: ", result)
			}
		}
	}
}
