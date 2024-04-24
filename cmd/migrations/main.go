package main

import (
	"database/sql"

	database "github.com/Damione1/thread-art-generator/pkg/db"
	"github.com/Damione1/thread-art-generator/pkg/util"
	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("👋 Starting migration")
	config, err := util.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to load config")
	}
	log.Info().Msg("👋 Config loaded")

	db, err := database.ConnectDb(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("👋 Failed to connect to database")
	}
	log.Info().Msg("👋 Connected to database")

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
				log.Info().Msg("🥝 No migration to run")
				return
			}

			version, dirty, errVersion := m.Version()
			if errVersion != nil {
				log.Error().Err(errVersion).Msg("🥝 Error retrieving version information")
				continue
			}

			if dirty {
				log.Warn().Err(err).Msg("🥝 Database in a dirty state")
				forceErr := m.Force(int(version))
				if forceErr != nil {
					log.Error().Err(forceErr).Msg("🥝 Failed to force version to clean state")
					continue
				}
				log.Info().Msg("🥝 Dirty state resolved by force. Retrying...")
				continue
			}

			log.Warn().Err(err).Msgf("🥝 Migration failed (attempt %d/%d)", i, retryCount)
			continue
		}
		log.Info().Msg("🥝 Migration ran successfully")
		return
	}

	log.Fatal().Msg("🥝 Failed to run migration after multiple attempts")
}
