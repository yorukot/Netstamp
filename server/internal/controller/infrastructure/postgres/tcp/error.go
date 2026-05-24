package pgtcp

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func mapTCPResultWriteError(err error) error {
	switch {
	case postgres.IsForeignKeyViolation(err, "fk_tcp_results_check"),
		postgres.IsForeignKeyViolation(err, "fk_tcp_results_probe"):
		return fmt.Errorf("tcp result invalid: %w", domaintcp.ErrInvalidResult)
	default:
		return err
	}
}
