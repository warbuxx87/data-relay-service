package util

import (
	"bitbucket.org/project-alkemi/data-relay-service/internal/app/data-relay-service/managers"
	"bitbucket.org/project-alkemi/dr-discovery-service-api"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type DataContext struct {
	Database            *gorm.DB
	Config              *Configuration
	RedisClient         *redis.Client
	DiscoveryServiceAPI *dr_discovery_service_api.DrDiscoveryServiceApi
	DataExchangeMgr     *managers.DataExchangeManager
	TLSServerManager    *managers.TLSServerManager
}
