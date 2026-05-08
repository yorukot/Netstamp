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
	httpListener, err := net.Listen("tcp", a.Config.HTTP.Addr)
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

	group.Go(func() error {
		<-groupCtx.Done()
		return a.shutdown()
	})

	return group.Wait()
}

func (a *Application) shutdown() error {
	a.Log.Info("application_stopping")

	ctx, cancel := context.WithTimeout(context.Background(), a.Config.ShutdownTimeout)
	defer cancel()

	var errs []error
	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown http: %w", err))
	}

	if a.DBPool != nil {
		a.DBPool.Close()
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
