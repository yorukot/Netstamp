package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (a *Application) Run(ctx context.Context) error {
	listenConfig := net.ListenConfig{}
	httpListener, err := listenConfig.Listen(ctx, "tcp", a.Config.HTTP.Addr)
	if err != nil {
		return fmt.Errorf("listen http: %w", err)
	}

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		a.Log.Info("http_server_started", zap.String("addr", httpListener.Addr().String()))
		err := a.HTTPServer.Serve(httpListener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	for _, worker := range a.Workers {
		if worker == nil {
			continue
		}
		group.Go(func() error {
			return worker.Run(groupCtx)
		})
	}

	group.Go(func() error {
		<-groupCtx.Done()
		return a.shutdown(groupCtx)
	})

	return group.Wait()
}

func (a *Application) shutdown(ctx context.Context) error {
	a.Log.Info("application_stopping")

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.Config.ShutdownTimeout)
	defer cancel()

	var errs []error
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown http: %w", err))
	}

	if a.DBPool != nil {
		a.DBPool.Close()
	}
	if a.Metrics != nil {
		if err := a.Metrics.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown metrics: %w", err))
		}
	}
	if a.Tracing != nil {
		if err := a.Tracing.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown tracing: %w", err))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	a.Log.Info("application_stopped")
	return nil
}
