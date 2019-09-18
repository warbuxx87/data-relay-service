package util

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Configuration struct {
	ListenIp   string
	ListenPort int

	SentryUrl string
	RedisIP   string
	RedisPort int

	DataRelayDiscoveryServiceUrl   string
	DataRelayDiscoveryRegisterKey  string
	DataRelayDiscoveryServiceAlias string

	TLSSeverIp           string
	TLSSeverPort         int
	TLSSeverCertLocation string
	TLSSeverKeyLocation  string
	// 0 = Info, 1 = Debug
	LogLevel int
}

type ActionResult struct {
}

func LoadConfigurationFile(filename string) (*Configuration, error) {
	configPath, err := resolveConfigPath(filename)
	if err != nil {
		return nil, err
	}
	log.Printf("Loading configuration from %s", configPath)
	configuration := Configuration{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &configuration); err != nil {
		return nil, err
	}
	return &configuration, nil
}

func resolveConfigPath(filename string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "configs", filename), nil
}

func ShowBanner() {

	fmt.Printf(`
    ___    ____                  _             ____  ____  _____
   /   |  / / /_____  ____ ___  (_)           / __ \/ __ \/ ___/
  / /| | / / //_/ _ \/ __  __ \/ /  ______   / / / / /_/ /\__ \
 / ___ |/ / ,< /  __/ / / / / / /  /_____/  / /_/ / _, _/___/ /
/_/  |_/_/_/|_|\___/_/ /_/ /_/_/           /_____/_/ |_|/____/
		`)

	fmt.Printf("- Alkemi Data Relay Service - V 0.0.1")
	fmt.Printf("\n")
	fmt.Printf("\n")
}
