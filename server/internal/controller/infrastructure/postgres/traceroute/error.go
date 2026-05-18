package pgtraceroute

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func mapTracerouteResultWriteError(err error) error {
	switch {
	case postgres.IsForeignKeyViolation(err, "fk_traceroute_results_check"),
		postgres.IsForeignKeyViolation(err, "fk_traceroute_results_probe"),
		postgres.IsForeignKeyViolation(err, "fk_traceroute_result_hops_result"):
		return fmt.Errorf("traceroute result invalid: %w", domaintraceroute.ErrInvalidResult)
	default:
		return err
	}
}
