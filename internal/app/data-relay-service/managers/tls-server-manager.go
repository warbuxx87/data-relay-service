package managers

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	common_message_models "bitbucket.org/project-alkemi/common-message-pkg/common-message-models"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"time"
)

type TLSServerManager struct {
	ConnectedClients *map[*net.Conn]bool
	RedisClient      *redis.Client
	ServiceUUID      string
}

func CreateTLSServerManager(serviceUUID string, redisClient *redis.Client) *TLSServerManager {

	tlsServerManager := TLSServerManager{}

	myMap := make(map[*net.Conn]bool)
	tlsServerManager.ConnectedClients = &myMap
	tlsServerManager.RedisClient = redisClient
	tlsServerManager.ServiceUUID = serviceUUID

	return &tlsServerManager
}

func (tlsServerManager *TLSServerManager) StartSever(tlsServerIP string, tlsServerPort int, tlsCertLocation string, tlsKeyLocation string) error {

	connectionCertificate, loadKeyPairError := tls.LoadX509KeyPair(tlsCertLocation, tlsKeyLocation)

	if loadKeyPairError != nil {
		logrus.Error(loadKeyPairError)
		return loadKeyPairError
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{connectionCertificate}}

	tlsListenConn, tlsListenError := tls.Listen("tcp", tlsServerIP+":"+strconv.Itoa(tlsServerPort), tlsConfig)

	if tlsListenError != nil {
		logrus.Info(tlsListenError)
		return tlsListenError
	}

	defer func() {
		tlsSeverCloseError := tlsListenConn.Close()

		if tlsSeverCloseError != nil {
			logrus.Error(tlsSeverCloseError)
		}
	}()

	fmt.Println("⇨ tls sever started on " + tlsServerIP + ":" + strconv.Itoa(tlsServerPort))

	for {

		clientConnection, acceptError := tlsListenConn.Accept()

		if acceptError != nil {
			logrus.Error(acceptError)
			continue
		}

		go HandleTLSConnection(tlsServerManager, &clientConnection)
	}

}

func HandleTLSConnection(tlsServerManager *TLSServerManager, conn *net.Conn) {

	defer func() {
		closeConnectionError := (*conn).Close()
		if closeConnectionError != nil {
			logrus.Error(closeConnectionError)
		}
	}()

	(*tlsServerManager.ConnectedClients)[conn] = true

	logrus.Info("Client Connected: " + (*conn).RemoteAddr().String())

	dec := gob.NewDecoder(*conn)

	for {

		var rawMessage common_message_models.RawMessageModel
		rawMessageUnMarshalError := dec.Decode(&rawMessage)

		if rawMessageUnMarshalError != nil {

			logrus.Info("Client Disconnected: " + (*conn).RemoteAddr().String())

			delete(*tlsServerManager.ConnectedClients, conn)

			return

		} else {

			if rawMessage.MessageType == common_message_consts.MessageType_PingMessage {

				logrus.Debug("Ping Message Received From: " + (*conn).RemoteAddr().String())
				tlsServerManager.TLSHandlePingMessage(conn, rawMessage)

			} else {

				// If not ping pong forward to all clients
				buffer := bytes.Buffer{}

				encodingError := gob.NewEncoder(&buffer).Encode(rawMessage)

				if encodingError != nil {
					logrus.Error(encodingError)
					raven.CaptureError(encodingError, nil)

				} else {

					publishError := tlsServerManager.RedisClient.Publish(tlsServerManager.ServiceUUID, base64.StdEncoding.EncodeToString(buffer.Bytes())).Err()

					if publishError != nil {
						logrus.Error(publishError)
						raven.CaptureError(publishError, nil)
					}
				}
			}
		}

	}

}

func (tlsServerManager *TLSServerManager) TLSHandlePingMessage(clientConn *net.Conn, message common_message_models.RawMessageModel) {

	if clientConn != nil {

		clientEnc := gob.NewEncoder(*clientConn)

		pongBuffer := bytes.Buffer{}

		pongMessage := common_message_models.PongMessageModel{TimeStamp: time.Now().UTC()}

		pongMessageMarshalError := gob.NewEncoder(&pongBuffer).Encode(pongMessage)

		if pongMessageMarshalError != nil {
			logrus.Error(pongMessageMarshalError)
			return
		}

		rawMessageModel := common_message_models.RawMessageModel{
			MessageType:     common_message_consts.MessageType_PongMessage,
			RawMessageBytes: pongBuffer.Bytes(),
			TimeStamp:       time.Now().UTC(),
		}

		writeDataToSocketError := clientEnc.Encode(rawMessageModel)

		if writeDataToSocketError != nil {
			logrus.Error(writeDataToSocketError)
			return
		}

	}
}
