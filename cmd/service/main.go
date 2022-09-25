package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DANDA322/user-balance-service/internal"
	"github.com/DANDA322/user-balance-service/internal/pgstore"
	"github.com/DANDA322/user-balance-service/internal/rest"
	"github.com/DANDA322/user-balance-service/pkg/converter"
	"github.com/DANDA322/user-balance-service/pkg/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

const addr = ":9999"

var (
	verbose    = lookupEnv("VERBOSE", "true")
	pgDSN      = lookupEnv("PG_DSN", "postgres://postgres:secret@localhost:5432/postgres")
	convURL    = lookupEnv("CONV_URL", "https://api.apilayer.com/exchangerates_data/latest?")
	convAPIKey = lookupEnv("CONV_API_KEY", "9lhi5Xm3MES5GPentAspvlOh5AX1VVPy")
)

func main() {
	log := logging.GetLogger(verbose)
	ctx := context.Background()
	store, err := pgstore.GetPGStore(ctx, log, pgDSN)
	if err != nil {
		log.Panicf("failed to get pg connection: %v", err)
	}
	conv := converter.NewConverter(convURL, convAPIKey)
	service := internal.NewApp(log, store, conv)
	router := rest.NewRouter(log, service)
	if err := startServer(ctx, log, router); err != nil {
		log.Panic("error: ", err)
	}
}

func startServer(ctx context.Context, log *logrus.Logger, r http.Handler) error {
	log.Info("Server start on ", addr)
	s := http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: time.Second * 30,
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		gcfCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		_ = s.Shutdown(gcfCtx)
		log.Info("Server is turned off")
	}()
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func lookupEnv(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
