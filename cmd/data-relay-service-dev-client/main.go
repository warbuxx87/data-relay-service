package main

import (
	common_message_consts "bitbucket.org/project-alkemi/common-message-pkg/common-message-consts"
	common_message_models "bitbucket.org/project-alkemi/common-message-pkg/common-message-models"
	"bytes"
	"encoding/gob"
	uuid "github.com/satori/go.uuid"
	"math/rand"
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

func main() {

	rand.Seed(time.Now().UnixNano())

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

		log.SetFlags(log.Lshortfile)

		cert, err := ioutil.ReadFile(config.TLSSeverCertLocation)

		if err != nil {
			logrus.Fatalf("Couldn't load file", err)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)

		conf := &tls.Config{
			RootCAs: certPool,
		}

		conn, err := tls.Dial("tcp", config.TLSSeverIp+":"+strconv.Itoa(config.TLSSeverPort), conf)
		if err != nil {
			logrus.Println(err)
			return
		}

		previousTime := time.Now().UTC()
		timeToTest := time.Now().UTC().Add(time.Duration(-3600) * time.Second)

		_ = previousTime
		_ = timeToTest

		logrus.Info(previousTime)
		logrus.Info(timeToTest)

		defer conn.Close()

		enc := gob.NewEncoder(conn)

		// Send price update
		zerp := common_message_models.OrderFilledMessageModel{
			TimeStamp:      time.Now().UTC(),
			OrderId:        uuid.NewV4().String(),
			Price:          float64(randomInt(0, 200)),
			BaseAsset:      "BTC",
			SecondaryAsset: "TRX",
			Amount:         float64(randomInt(1, 100)),
			IsBuy:          true,
		}

		buffer := bytes.Buffer{}

		gob.NewEncoder(&buffer).Encode(zerp)

		herpDerp := common_message_models.RawMessageModel{
			TimeStamp:       time.Now().UTC(),
			UUID:            uuid.NewV4().String(),
			AccessUUID:      "0fb4a0a6-0a96-4d2d-af62-4d09c2208c85",
			AccessToken:     "e93b8d10-7880-41ab-a602-5595ba197c44",
			MessageType:     common_message_consts.MessageType_OrderFilledMessage,
			RawMessageBytes: buffer.Bytes()}

		errEnc := enc.Encode(herpDerp)

		if errEnc != nil {
			logrus.Error(errEnc)
		}

		if errEnc != nil {
			logrus.Error(errEnc)
		}

		for {

			dec := gob.NewDecoder(conn)

			var rawMessage common_message_models.RawMessageModel
			rawMessageUnMarshalError := dec.Decode(&rawMessage)

			if rawMessageUnMarshalError != nil {
				logrus.Error(rawMessageUnMarshalError)
			}

			logrus.Info(rawMessage.MessageType)

			//logrus.Println(rawMessage.MessageType)
			time.Sleep(time.Duration(150) * time.Millisecond)

			// Send price update
			zerp := common_message_models.OrderFilledMessageModel{
				TimeStamp:      time.Now().UTC(),
				OrderId:        uuid.NewV4().String(),
				Price:          float64(randomInt(0, 200)),
				BaseAsset:      "BTC",
				SecondaryAsset: "TRX",
				Amount:         float64(randomInt(1, 100)),
				IsBuy:          true,
			}

			buffer := bytes.Buffer{}

			gob.NewEncoder(&buffer).Encode(zerp)

			herpDerp := common_message_models.RawMessageModel{
				TimeStamp:       time.Now().UTC(),
				UUID:            uuid.NewV4().String(),
				AccessUUID:      "0fb4a0a6-0a96-4d2d-af62-4d09c2208c85",
				AccessToken:     "e93b8d10-7880-41ab-a602-5595ba197c44",
				MessageType:     common_message_consts.MessageType_OrderFilledMessage,
				RawMessageBytes: buffer.Bytes()}

			errEnc := enc.Encode(herpDerp)

			if errEnc != nil {
				logrus.Error(errEnc)
			}
		}

	}
}
