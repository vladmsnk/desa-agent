package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"desa-agent/internal/adapters"
	"desa-agent/internal/config"
	"desa-agent/internal/transport"
	"desa-agent/internal/usecase"
)

type App struct {
	cfg        *config.Config
	grpcServer *grpc.Server
	idp        adapters.IdentityProvider
	logger     *slog.Logger
}

func New(cfg *config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	idp, err := adapters.NewIdentityProvider(cfg.IDP)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity provider: %w", err)
	}

	logger.Info("identity provider adapter created",
		"type", cfg.IDP.Type,
		"host", cfg.IDP.Host,
	)

	usersUC := usecase.NewUsersUseCase(idp)

	grpcServer := grpc.NewServer()

	usersService := transport.NewUsersServiceServer(usersUC)
	usersService.Register(grpcServer)

	reflection.Register(grpcServer)

	return &App{
		cfg:        cfg,
		grpcServer: grpcServer,
		idp:        idp,
		logger:     logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	listener, err := net.Listen("tcp", a.cfg.GRPC.Address())
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", a.cfg.GRPC.Address(), err)
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("starting gRPC server", "address", a.cfg.GRPC.Address())
		if err := a.grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("context canceled, shutting down")
	case sig := <-sigCh:
		a.logger.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		return err
	}

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting down application")

	a.grpcServer.GracefulStop()

	if err := a.idp.Close(); err != nil {
		a.logger.Error("failed to close identity provider", "error", err)
	}

	a.logger.Info("application shutdown complete")
	return nil
}
