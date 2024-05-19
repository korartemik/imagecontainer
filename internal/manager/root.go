package manager

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"os"
	"tagestest/internal/config"
	"tagestest/internal/grpc/image_container"
	"tagestest/internal/lib/clock/real_clock"
	"tagestest/internal/storage"
)

type Root struct {
	debugServer *http.Server
	errorChan   chan error
	gRPCServer  *grpc.Server
	db          *badger.DB
	cfg         *config.Config
}

func NewRoot() *Root {
	return &Root{}
}

func (r *Root) Register(ctx context.Context) error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger()

	var err error
	r.cfg, err = config.New()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return err
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error().Msg("Recovered from panic")

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
	))

	db, err := badger.Open(badger.DefaultOptions(r.cfg.Badger.Path))
	if err != nil {
		return err
	}
	str := storage.NewImageStorage(db, real_clock.NewRealClock())

	r.db = db

	image_container.Register(gRPCServer, str, r.cfg)

	return nil
}

func (r *Root) Resolve(ctx context.Context, shutdown chan os.Signal) os.Signal {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", r.cfg.Server.Port))
	if err != nil {
		log.Error().Err(err)
		return os.Interrupt
	}

	log.Info().Msg("grpc server started")

	// Запускаем обработчик gRPC-сообщений
	if err := r.gRPCServer.Serve(l); err != nil {
		log.Error().Err(err)
		return os.Interrupt
	}

	for {
		select {
		case err := <-r.errorChan:
			log.Err(err).Msg("error occurred")

		case <-ctx.Done():
			log.Info().Msg("context done")
			return os.Interrupt

		//	заканчиваем работу
		case sig := <-shutdown:
			return sig
		}
	}
}

func (r *Root) Release(signal os.Signal) {
	log.Info().Msgf("shutdown started with signal : [%d]", signal)
	defer log.Info().Msg("shutdown completed")

	r.gRPCServer.GracefulStop()

	if err := r.db.Close(); err != nil {
		log.Err(err).Msg("could not close db gracefully")
	}
}
