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
	LogLevel             int
	HealthCheckFrequency int
}

type ClientConfiguration struct {
	TLSSeverIp           string
	TLSSeverPort         int
	TLSSeverCertLocation string
	TLSSeverKeyLocation  string
	// 0 = Info, 1 = Debug
	LogLevel    int
	AccessUUID  string
	AccessToken string
}

type ActionResult struct {
}

func LoadConfigurationFile(filename string) (*Configuration, error) {
	configPath, err := resolveConfigPath(filename)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(configPath); err != nil {
		if err := generateDefaultConfig(filename); err != nil {
			return nil, err
		}
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

func LoadClientConfiguration(filename string) (*ClientConfiguration, error) {
	configPath, err := resolveConfigPath(filename)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(configPath); err != nil {
		if err := generateClientConfig(configPath); err != nil {
			return nil, err
		}
	}
	log.Info("Loading client configuration from %s:", configPath)
	config := &ClientConfiguration{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func generateClientConfig(configPath string) error {
	cfg := defaultClientConfig()
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, data, 0664)
}

func resolveConfigPath(filename string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "configs", filename), nil
}

func generateDefaultConfig(filename string) error {
	configPath, err := resolveConfigPath(filename)
	if err != nil {
		return err
	}
	cfg := defaultConfig()
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, data, 0664)
}

func defaultConfig() *Configuration {
	return &Configuration{
		ListenIp:                       "127.0.0.1",
		ListenPort:                     7070,
		SentryUrl:                      "",
		RedisIP:                        "localhost",
		RedisPort:                      6379,
		TLSSeverIp:                     "127.0.0.1",
		TLSSeverPort:                   7771,
		TLSSeverCertLocation:           "certs/devcert.crt",
		TLSSeverKeyLocation:            "certs/devcert.key",
		DataRelayDiscoveryServiceUrl:   "http://localhost:7060",
		DataRelayDiscoveryRegisterKey:  "registerme",
		DataRelayDiscoveryServiceAlias: "SomeAlias",
		LogLevel:                       1,
	}
}

func defaultClientConfig() *ClientConfiguration {
	return &ClientConfiguration{
		TLSSeverIp:           "127.0.0.1",
		TLSSeverPort:         7771,
		TLSSeverCertLocation: "certs/devcert.crt",
		TLSSeverKeyLocation:  "certs/devcert.key",
		LogLevel:             1,
		AccessToken:          "",
		AccessUUID:           "",
	}
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
