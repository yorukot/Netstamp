package pgping

import (
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func sqlcPingStatus(value domainping.Status) sqlc.PingStatus {
	return sqlc.PingStatus(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) sqlc.NullIpFamily {
	if value == nil {
		return sqlc.NullIpFamily{}
	}

	return sqlc.NullIpFamily{
		IpFamily: sqlc.IpFamily(*value),
		Valid:    true,
	}
}
