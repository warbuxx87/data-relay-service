package main

import (
	"bitbucket.org/project-alkemi/common-pkg/gen/consts"
	"bitbucket.org/project-alkemi/common-pkg/gen/messages"
	"bitbucket.org/project-alkemi/common-pkg/gen/services"
	"bufio"
	"context"
	"errors"
	"flag"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	conn   *grpc.ClientConn
	client *services.DataRelayServiceClient
}

func NewDevClient(config *util.ClientConfiguration) *DevClient {
	return &DevClient{config: config}
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

	conn, err := grpc.Dial(c.config.TLSSeverIp+":"+strconv.Itoa(c.config.TLSSeverPort), grpc.WithTransportCredentials(credentials.NewTLS(conf)))

	if err != nil {
		return err
	}

	derp := services.NewDataRelayServiceClient(conn)

	c.conn = conn
	c.client = &derp

	return nil
}

func (c *DevClient) close() error {
	if c.client == nil {
		return errors.New("connection is closed")
	}
	return c.conn.Close()
}

func (c *DevClient) sendOrderFilled() error {

	// Send order filled
	currentTime := time.Now().UTC()

	currentTimeMessage := messages.Timestamp{
		Seconds: currentTime.Unix(),
		Nanos:   int32(currentTime.Nanosecond()),
	}

	newOrderFilledMessage := messages.OrderFilled{
		UpdateTime:     &currentTimeMessage,
		OrderId:        uuid.NewV4().String(),
		Price:          float32(randomInt(0, 200)),
		BaseAsset:      "BTC",
		SecondaryAsset: "USDT",
		Amount:         float32(randomInt(1, 100)),
		IsBuy:          true,
	}

	toSendMessage := messages.AlkemiMessage{
		UUID:         uuid.NewV4().String(),
		AccessUUID:   c.config.AccessUUID,
		AccessToken:  c.config.AccessToken,
		SentTime:     &currentTimeMessage,
		MessageType:  consts.MessageType_ORDER_FILLED,
		OrderFilledM: &newOrderFilledMessage,
	}

	sendOrderFilledResponse, sendOrderFilledResponseError := (*c.client).SendData(context.Background(), &toSendMessage)

	if sendOrderFilledResponse != nil {
		if sendOrderFilledResponse.Success == false {
			logrus.Error(sendOrderFilledResponse.ErrorStr)
		}

		logrus.Info("Send order filled success")
	}

	return sendOrderFilledResponseError
}

func (c *DevClient) sendPriceUpdate() error {

	// Send order filled
	currentTime := time.Now().UTC()

	currentTimeMessage := messages.Timestamp{
		Seconds: currentTime.Unix(),
		Nanos:   int32(currentTime.Nanosecond()),
	}

	newPriceUpdateMessage := messages.PriceUpdate{
		UUID:           uuid.NewV4().String(),
		UpdateTime:     &currentTimeMessage,
		BaseAsset:      "BTC",
		SecondaryAsset: "USDT",
		Price:          float32(randomInt(0, 200)),
	}

	toSendMessage := messages.AlkemiMessage{
		UUID:         uuid.NewV4().String(),
		AccessUUID:   c.config.AccessUUID,
		AccessToken:  c.config.AccessToken,
		SentTime:     &currentTimeMessage,
		MessageType:  consts.MessageType_PRICE_UPDATE,
		PriceUpdateM: &newPriceUpdateMessage,
	}

	sendPriceUpdateResponse, sendPriceUpdateError := (*c.client).SendData(context.Background(), &toSendMessage)

	if sendPriceUpdateResponse != nil {
		if sendPriceUpdateResponse.Success == false {
			logrus.Error(sendPriceUpdateResponse.ErrorStr)
		} else {
			logrus.Info("Send price update success")
		}
	}

	return sendPriceUpdateError
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
	log.Println("1. PriceUpdateMessage")
	log.Println("2. OrderFilledMessage")
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
			priceUpdateError := c.sendPriceUpdate()
			if priceUpdateError != nil {
				logrus.Error(priceUpdateError)
			}
		case "2":
			orderFilledError := c.sendOrderFilled()
			if orderFilledError != nil {
				logrus.Error(orderFilledError)
			}
		default:
			break
		}

	}
}
