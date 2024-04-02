package main

import (
	"database/sql"

	database "github.com/Damione1/thread-art-generator/pkg/db"
	"github.com/Damione1/thread-art-generator/pkg/util"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}

	db, err := database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}

	RunDBMigration(&config, db)
}

func RunDBMigration(config *util.Config, db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("🥝 Failed to create migration driver")
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", config.PostgresDb, driver)
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
		log.Info().Msg("🥝 Migration ran successfully")
		return
	}

	log.Fatal().Msg("🥝 Failed to run migration after multiple attempts")
}
