package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"log"
	"os"
	"time"
)

type PostgresRepository struct {
	database *sql.DB
}

func MakePostgresRepository() Repository {
	var err error
	config.DB, err = sql.Open("pgx", config.AppConfig.DatabaseDsn)

	if err != nil {
		fmt.Println("Unable to connect to database.")
		os.Exit(1)
	} else {
		fmt.Println("Database created.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := config.DB.PingContext(ctx); err != nil {
		fmt.Println("Unable to connect to database.")
		os.Exit(1)
	} else {
		fmt.Println("Database connected.")
	}

	var repository = PostgresRepository{database: config.DB}

	_, seedError := repository.database.Exec(`
CREATE TABLE IF NOT EXISTS "url" (
	"id" VARCHAR(12) NOT NULL,
	"short" VARCHAR(128) NOT NULL,
	"original" VARCHAR(128) NOT NULL UNIQUE,
	"user_id" VARCHAR(12) NULL DEFAULT NULL,
	PRIMARY KEY ("id")
);`)

	if seedError != nil {
		log.Fatal(seedError)
	}

	return repository
}

func (r PostgresRepository) Insert(url URL) (URL, error) {
	var dbURL = URL{}

	err := r.database.QueryRow(`SELECT id, original, short, user_id FROM url WHERE id = $1 OR original= $2;`, url.ID, url.Original).Scan(&dbURL.ID, &dbURL.Original, &dbURL.Short, &dbURL.UserID)
	if err != nil {
		_, e := r.database.Exec(`INSERT INTO url (id, short, original, user_id) VALUES ($1, $2, $3, $4);`, url.ID, url.Short, url.Original, url.UserID)
		if e != nil {
			return url, e
		}
		return url, nil
	}

	return dbURL, ErrorURLExists
}

func (r PostgresRepository) Get(id string) (URL, bool) {
	var url URL

	err := r.database.QueryRow(`SELECT id, short, original, user_id FROM url WHERE id=$1`, id).Scan(&url.ID, &url.Short, &url.Original, &url.UserID)
	if err != nil {
		fmt.Println(err)
		return url, false
	}

	return url, true
}

func (r PostgresRepository) GetAllByUser(userID string) ([]URL, error) {
	var count int
	e := r.database.QueryRow(`SELECT COUNT(*) FROM url WHERE user_id=$1`, userID).Scan(&count)
	if e != nil {
		return nil, e
	}

	rows, err := r.database.Query(`SELECT id, short, original, user_id FROM url WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	urls := make([]URL, count)

	i := 0

	for rows.Next() {
		var url URL
		err := rows.Scan(&url.ID, &url.Short, &url.Original, &url.UserID)
		if err != nil {
			return nil, err
		}

		urls[i] = url
		i++
	}

	return urls[:i], nil
}
