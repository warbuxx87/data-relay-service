package util

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/tkanos/gonfig"
)

type Configuration struct {

	ListenIp   string
	ListenPort int

	SentryUrl string

	// 0 = Info, 1 = Debug
	LogLevel int
}

type ActionResult struct {

}

func LoadConfigurationFile() (*Configuration, error) {

	configuration := Configuration{}
	err := gonfig.GetConf("config.json", &configuration)

	if err != nil {
		logrus.WithFields(logrus.Fields{}).Error(err)
		return nil, errors.New("Failed to load configuration file")
	}

	return &configuration, nil
}

func ShowBanner() {

	fmt.Printf(`
    ___    ____                  _             ____  ____ 
   /   |  / / /_____  ____ ___  (_)           / __ \/ __ \
  / /| | / / //_/ _ \/ __  __ \/ /  ______   / / / / /_/ /
 / ___ |/ / ,< /  __/ / / / / / /  /_____/  / /_/ / _, _/
/_/  |_/_/_/|_|\___/_/ /_/ /_/_/           /_____/_/ |_|
		`)


	fmt.Printf("- Alkemi Data Relay - V 0.0.1")
	fmt.Printf("\n")
	fmt.Printf("\n")
}
