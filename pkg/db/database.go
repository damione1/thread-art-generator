package database

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/pkg/util"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func ConnectDb(config *util.Config) (*sql.DB, error) {
	log.Log().Msg("🥝 Connecting to database..." + config.PostgresUser + " " + config.PostgresPassword + " " + config.PostgresDb)
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
