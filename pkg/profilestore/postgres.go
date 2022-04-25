package profilestore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gerladeno/authorization-service/pkg/common"
	"github.com/jackc/pgconn"

	"github.com/gerladeno/authorization-service/pkg/metrics"
	"github.com/gerladeno/authorization-service/pkg/models"
	"github.com/jackc/pgx/v4"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

type PG struct {
	db      *pgx.Conn
	dsn     string
	log     *logrus.Entry
	metrics *metrics.DBClient
}

func GetPGStore(ctx context.Context, log *logrus.Logger, dsn string) (*PG, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.PreferSimpleProtocol = true
	db, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(ctx); err != nil {
		return nil, err
	}
	fn := func() float64 {
		return 1.0
	}
	return &PG{
		db:      db,
		dsn:     dsn,
		log:     log.WithField("module", "profileStore"),
		metrics: metrics.NewDBClient(config.Database, config.Host, fmt.Sprintf("%d", config.Port), fn).AutoRegister(),
	}, nil
}

func (pg *PG) Migrate(direction migrate.MigrationDirection) error {
	conn, err := sql.Open("pgx", pg.dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			pg.log.Error("err closing migration connection")
		}
	}()
	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, er := migrations.ReadDir(path)
			if er != nil {
				return nil, er
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}
	_, err = migrate.Exec(conn, "postgres", asset, direction)
	return err
}

func (pg *PG) GetUser(ctx context.Context, phone string) (*models.User, error) {
	query := fmt.Sprintf(`SELECT uuid, phone, created, updated
FROM user_model
WHERE phone = '%s';`, phone)
	var started time.Time
	var err error
	var result models.User
	for i := 0; i < common.GlobalRequestRetries; i++ {
		started = time.Now()
		err = pgxscan.Get(ctx, pg.db, &result, query)
		switch {
		case err == nil:
		case errors.Is(err, pgx.ErrNoRows):
			return nil, common.ErrPhoneNotFound
		default:
			pg.metrics.ErrsTotal.WithLabelValues("GetUser").Inc()
			continue
		}
		pg.metrics.TimeTotal.WithLabelValues("GetUser").Add(time.Since(started).Seconds())
		return &result, nil
	}
	pg.log.Debugf("err selecting from pg: %s", err)
	return nil, err
}

func (pg *PG) UpsertUser(ctx context.Context, user *models.User) error {
	query := `
INSERT INTO user_model (uuid, phone, created, updated)
VALUES ($1, $2, $3, $4)
ON CONFLICT (phone) DO UPDATE SET phone     = excluded.phone,
                                 updated    = NOW()
;`
	var started time.Time
	var err error
	var result pgconn.CommandTag
	for i := 0; i < common.GlobalRequestRetries; i++ {
		started = time.Now()
		result, err = pg.db.Exec(
			ctx, query, user.UUID, user.Phone,
			time.Now().UTC().Format(common.PGDatetimeFmt), time.Now().UTC().Format(common.PGDatetimeFmt))
		if err != nil {
			pg.metrics.ErrsTotal.WithLabelValues("UpsertUser").Inc()
			continue
		}
		if result.RowsAffected() == 0 {
			err = errors.New("err user not upserted")
			pg.metrics.ErrsTotal.WithLabelValues("UpsertUser").Inc()
			continue
		}
		pg.metrics.TimeTotal.WithLabelValues("UpsertUser").Add(time.Since(started).Seconds())
		return nil
	}
	err = fmt.Errorf("err inserting to pg: %w", err)
	pg.log.Debug(err)
	return err
}
