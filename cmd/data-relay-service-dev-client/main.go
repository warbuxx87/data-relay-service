package main

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	common_message_models "bitbucket.org/project-alkemi/common-message-pkg/common-message-models"
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/util"
	"crypto/tls"
	"crypto/x509"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func init() {

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

type DevClient struct {
	config *util.ClientConfiguration
	conn *tls.Conn
}

func NewDevClient(config *util.ClientConfiguration) *DevClient {
	return &DevClient{config:config}
}

func (c *DevClient) connect() error {
	cert, err := ioutil.ReadFile(c.config.TLSSeverCertLocation)
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	conf := &tls.Config{
		RootCAs: certPool,
	}
	conn, err := tls.Dial("tcp", c.config.TLSSeverIp+":"+strconv.Itoa(c.config.TLSSeverPort), conf)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *DevClient) close() error {
	if c.conn == nil {
		return errors.New("connection is closed")
	}
	return c.conn.Close()
}

func (c *DevClient) sendSubscriptionMessage() error {
	subscriptionMessage := common_message_models.SubscribeDataChannelMessageModel{
		DataChannelType: common_message_consts.DataChannelType_AllData,
	}
	return c.sendMessage(subscriptionMessage, common_message_consts.MessageType_UnSubscribeDataChannelMessage)
}

func (c *DevClient) sendOrderFilled() error {
	// Send price update
	orderFilledMessage := common_message_models.OrderFilledMessageModel{
		TimeStamp:      time.Now().UTC(),
		OrderId:        uuid.NewV4().String(),
		Price:          float64(randomInt(0, 200)),
		BaseAsset:      "BTC",
		SecondaryAsset: "USDT",
		Amount:         float64(randomInt(1, 100)),
		IsBuy:          true,
	}
	return c.sendMessage(orderFilledMessage, common_message_consts.MessageType_OrderFilledMessage)
}

func (c *DevClient) sendPriceUpdate() error {
	priceUpdateMessage := common_message_models.PriceUpdateMessageModel{
		TimeStamp:      time.Now().UTC(),
		UUID:        uuid.NewV4().String(),
		Price:          float64(randomInt(0, 200)),
		BaseAsset:      "BTC",
		SecondaryAsset: "USDT",
	}
	return c.sendMessage(priceUpdateMessage, common_message_consts.MessageType_PriceUpdateMessage)
}

func (c *DevClient) sendPingMessage() error {
	pingMessage := common_message_models.PingMessageModel{
		TimeStamp:      time.Now().UTC(),
	}
	return c.sendMessage(pingMessage, common_message_consts.MessageType_PingMessage)
}

func (c *DevClient) sendUnsubscribeMessage() error {
	subscriptionMessage := common_message_models.SubscribeDataChannelMessageModel{
		DataChannelType: common_message_consts.DataChannelType_AllData,
	}
	return c.sendMessage(subscriptionMessage, common_message_consts.MessageType_SubscribeDataChannelMessage)
}

func (c *DevClient) sendMessage(message interface{}, typeConst common_message_consts.MessageTypeConst) error {
	log.Println("Sending message ", typeConst)
	buffer := bytes.Buffer{}
	encodeError := gob.NewEncoder(&buffer).Encode(message)
	if encodeError != nil {
		return encodeError
	}
	rawMessage := common_message_models.RawMessageModel{
		TimeStamp:       time.Now().UTC(),
		UUID:            uuid.NewV4().String(),
		AccessUUID: 	c.config.AccessUUID,
		AccessToken: 	c.config.AccessToken,
		MessageType:     typeConst,
		RawMessageBytes: buffer.Bytes(),
	}
	networkEncodeError := gob.NewEncoder(c.conn).Encode(rawMessage)
	if networkEncodeError != nil {
		return networkEncodeError
	}
	return nil
}

func (c *DevClient) readResponse() (*common_message_models.RawMessageModel, error) {
	dec := gob.NewDecoder(c.conn)
	var rawMessage common_message_models.RawMessageModel
	rawMessageUnMarshalError := dec.Decode(&rawMessage)
	if rawMessageUnMarshalError != nil {
		return nil, rawMessageUnMarshalError
	}
	if rawMessage.FromDataRelay == true {
		logrus.Info("Message From Relay : "+rawMessage.FromDataRelayAlias+" "+rawMessage.UUID+" "+rawMessage.TimeStamp.String()+" ", rawMessage.MessageType)
	} else {
		logrus.Info("Message: "+rawMessage.UUID+" "+rawMessage.TimeStamp.String()+" ", rawMessage.MessageType)
	}
	return &rawMessage, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	prod := flag.Bool("prod", false, "Use '$go run ./cmd/... -prod' to set to true, changing default configuration file from 'dev.json' to 'config.json'")
	flag.Parse()

	var configFilename string
	if *prod {
		configFilename = "client_config.json"
	} else {
		configFilename = "dev_client_config.json"
	}
	config, configLoadError := util.LoadClientConfiguration(configFilename)
	if configLoadError != nil {
		logrus.Fatalf("failed to parse config: %v", configLoadError)
	}
	// Set the logrus config level
	if config.LogLevel == 0 {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Info("Log Level: Info")
	} else {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("Log Level: Debug")

	}

	log.SetFlags(log.Lshortfile)

	previousTime := time.Now().UTC()
	timeToTest := time.Now().UTC().Add(time.Duration(-3600) * time.Second)

	_ = previousTime
	_ = timeToTest

	logrus.Info(previousTime)
	logrus.Info(timeToTest)

	client := NewDevClient(config)
	connectionError := client.connect()
	if connectionError != nil {
		logrus.Fatalf("failed to connect to server: %v", connectionError)
	}
	defer func() {
		if closeError := client.close(); closeError != nil {
			logrus.Println(closeError)
		}
	}()
	client.processInput()
}

func (c *DevClient) outputOptions() {
	log.Println("> Select one of the options below to test data-relay-service.")
	log.Println("1. Ping Service")
	log.Println("2. PriceUpdateMessage")
	log.Println("3. OrderFilledMessage")
	log.Println("4. Subscribe to Data Channel")
	log.Println("5. Unsubscribe From Data Channel")
	log.Println("Press q to quit")
	log.Println(">")
}

func (c *DevClient) processInput() {
	c.outputOptions()
	reader := bufio.NewReader(os.Stdin)
	for {
		input, inputError := reader.ReadString('\n')
		if inputError != nil {
			log.Println("Invalid input. Please try again")
			continue
		}
		if strings.TrimSpace(strings.ToLower(input)) == "q" {
			log.Println("Thank you!")
			break
		}
		switch strings.TrimSpace(input) {
		case "1":
			pingError := c.sendPingMessage()
			if pingError != nil {
				log.Println(pingError)
			}
			break
		case "2":
			priceUpdateError := c.sendPriceUpdate()
			if priceUpdateError != nil {
				log.Println(priceUpdateError)
			}
		case "3":
			orderFilledError := c.sendOrderFilled()
			if orderFilledError != nil {
				log.Println(orderFilledError)
			}
		case "4":
			subscribeError := c.sendSubscriptionMessage()
			if subscribeError != nil {
				log.Println(subscribeError)
			}
		case "5":
			unSubscribeError := c.sendUnsubscribeMessage()
			if unSubscribeError != nil {
				log.Println(unSubscribeError)
			}
		default:
			break
		}
		rawMessage, err := c.readResponse()
		if err != nil {
			log.Println(err)
		}else {
			if rawMessage.FromDataRelay == true {
				logrus.Info("Message From Relay : "+rawMessage.FromDataRelayAlias+" "+rawMessage.UUID+" "+rawMessage.TimeStamp.String()+" ", rawMessage.MessageType)
			} else {
				logrus.Info("Message: "+rawMessage.UUID+" "+rawMessage.TimeStamp.String()+" ", rawMessage.MessageType)
			}
		}
	}
}