package database

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/pkg/util"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func ConnectDb(config *util.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=db user=%s password=%s dbname=%s TimeZone=America/New_York sslmode=disable",
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDb,
	))
	if err != nil {
		log.Fatal().Err(err).Msg("🥝 Failed to connect to database")
	}

	boil.SetDB(db)

	config.DB = db

	return db, nil
}

func RunDBMigration(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("🥝 Failed to create migration driver")
	}

	m, err := migrate.NewWithDatabaseInstance("file://../migrations", "postgres", driver)
	if err != nil {
		log.Fatal().Err(err).Msg("🥝 Failed to create migration instance")
	}

	retryCount := 3
	for i := 1; i <= retryCount; i++ {
		err = m.Up()
		if err != nil {
			if err == migrate.ErrNoChange {
				log.Print("🥝 No migration to run")
				return
			} else {
				log.Warn().Err(err).Msgf("🥝 Migration failed (attempt %d/%d)", i, retryCount)
				if revertErr := m.Down(); revertErr != nil {
					log.Error().Err(revertErr).Msg("🥝 Failed to revert migration")
				}
				continue
			}
		}
		// Migration succeeded
		return
	}

	log.Fatal().Msg("🥝 Failed to run migration after multiple attempts")
}
