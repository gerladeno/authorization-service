package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gerladeno/authorization-service/pkg/common"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/gerladeno/authorization-service/pkg/profilestore"

	"github.com/gerladeno/authorization-service/pkg/authentication"
	"github.com/gerladeno/authorization-service/pkg/authorization"
	"github.com/gerladeno/authorization-service/pkg/rest"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const httpPort = 3000

var version = `0.0.0`

func main() {
	log := GetLogger(true)
	log.Infof("starting authorization service version %s", version)
	if err := godotenv.Load(); err != nil {
		if common.RunsInContainer() {
			log.Infof("running in container, no .env file: %v", err)
		} else {
			log.Panic(err)
		}
	}

	var (
		flashCallHost   = os.Getenv("FLASHCALL_HOST")
		flashCallID     = os.Getenv("FLASHCALL_ID")
		flashCallSecret = os.Getenv("FLASHCALL_SECRET")
		signingKey      = os.Getenv("PRIVATE_SIGNING_KEY")
		host            = "localhost"
		pgDSN           = os.Getenv("PG_DSN")
	)
	if common.RunsInContainer() {
		pgDSN = strings.ReplaceAll(pgDSN, "localhost:5433", "auth_pg:5432")
	}
	ctx := context.Background()
	pg, err := profilestore.GetPGStore(ctx, log, pgDSN)
	if err != nil {
		panic(fmt.Errorf("err connecting to pg: %w", err))
	}
	if err = pg.Migrate(migrate.Up); err != nil {
		panic(fmt.Errorf("err migrating pg: %w", err))
	}
	flashcall := authentication.New(log, flashCallHost, flashCallID, flashCallSecret)
	auth := authorization.New(log, flashcall, signingKey)
	router := rest.NewRouter(log, auth, pg, host, version)
	if err = startServer(ctx, router, log); err != nil {
		log.Fatal(err)
	}
}

func startServer(ctx context.Context, router http.Handler, log *logrus.Logger) error {
	log.Infof("starting server on port %d", httpPort)
	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		Handler:           router,
	}
	errCh := make(chan error)
	go func() {
		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	select {
	case err := <-errCh:
		return err
	case <-sigCh:
	}
	log.Info("terminating...")
	gfCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.Shutdown(gfCtx)
}

func GetLogger(verbose bool) *logrus.Logger {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})
	if verbose {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("log level set to debug")
	}
	return log
}
