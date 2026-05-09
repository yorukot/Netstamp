package pgping

import (
	"fmt"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
)

func mapPingResultWriteError(err error) error {
	switch {
	case postgres.IsForeignKeyViolation(err, "fk_ping_results_project_check"),
		postgres.IsForeignKeyViolation(err, "fk_ping_results_project_probe"):
		return fmt.Errorf("ping result invalid: %w", domainping.ErrInvalidResult)
	default:
		return err
	}
}
