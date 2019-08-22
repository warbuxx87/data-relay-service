package util

import (
	"bitbucket.org/project-alkemi/dr-discovery-service-api"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type DataContext struct {
	Database            *gorm.DB
	Config              *Configuration
	Redis               *redis.Client
	DiscoveryServiceAPI *dr_discovery_service_api.DrDiscoveryServiceApi
}
