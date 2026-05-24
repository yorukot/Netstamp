package pgtcp

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func sqlcTCPStatus(value domaintcp.Status) sqlc.TcpStatus {
	return sqlc.TcpStatus(value)
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

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
