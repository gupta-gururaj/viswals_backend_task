package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/jmoiron/sqlx"
	 _ "github.com/golang-migrate/migrate/source/file"

)

const (
	connectAttempts = 10
	connectTimeout  = time.Second
	maxOpenConns    = 5
)

var (
	ErrDBNotInitialized = errors.New("database is not initialized")
	ErrNoData          = errors.New("requested data does not exist")
	ErrDuplicate       = errors.New("data to create already exists")


)

type Postgres struct {
	DB *sqlx.DB
}

func New(connectionStr string) (Postgres, error) {
	var (
		db        *sql.DB
		migration *migrate.Migrate
		driver    database.Driver
		err       error
	)

	for attempt := 0; attempt < connectAttempts; attempt++ {

		db, err = sql.Open("postgres", connectionStr)
		if err == nil {
			break
		}
		time.Sleep(connectTimeout)
	}
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to connect to database: %w", err)
	}


	driver, err = postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return Postgres{}, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migration, err = migrate.NewWithDatabaseInstance(
		"file://./migration",
		"postgres", driver)
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to create migrations instance: %w", err)
	}
	// Check the current version
	version, dirty, err := migration.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return Postgres{}, fmt.Errorf("failed to get migration version: %w", err)
	}

	// If dirty (failed migration), force reset to last known good state
	if dirty {
		if err := migration.Force(int(version)); err != nil {
			return Postgres{}, fmt.Errorf("failed to force migration reset: %w", err)
		}
	}

	// if err := migration.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
	// 	return Postgres{},err
	// }
	
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return Postgres{}, fmt.Errorf("failed to apply migrations: %w", err)
	}

	sdb := sqlx.NewDb(db, "postgres")
	sdb.SetMaxOpenConns(maxOpenConns)

	err = sdb.Ping()
	if err != nil {
		return Postgres{}, fmt.Errorf("failed to ping database: %w", err)
	}

	return Postgres{
		DB: sdb,
	}, nil
}
