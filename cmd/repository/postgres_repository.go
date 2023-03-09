package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/LorezV/url-shorter.git/cmd/config"
	"github.com/jackc/pgerrcode"
	"log"
	"os"
	"strings"
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

	_, seedError := repository.database.ExecContext(context.Background(), `
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

func (r PostgresRepository) Insert(context context.Context, url URL) (URL, error) {
	_, err := r.database.ExecContext(context, `INSERT INTO url (id, short, original, user_id) VALUES ($1, $2, $3, $4);`, url.ID, url.Short, url.Original, url.UserID)

	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			var dbURL URL
			err := r.database.QueryRowContext(context, `SELECT id, short, original, user_id FROM url WHERE id=$1 OR original=$2;`, url.ID, url.Original).Scan(&dbURL.ID, &dbURL.Short, &dbURL.Original, &dbURL.UserID)
			if err != nil {
				return url, err
			}

			return dbURL, ErrorUrlDuplicate
		}
		return url, err
	}

	return url, nil
}

func (r PostgresRepository) InsertMany(context context.Context, urls []URL) ([]URL, error) {
	tx, err := r.database.Begin()
	if err != nil {
		return urls, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(context, `
		INSERT INTO url (id, short, original, user_id) VALUES ($1, $2, $3, $4)
		ON CONFLICT(original) DO UPDATE SET original=$3 
		RETURNING id, short, original, user_id;
	`)
	if err != nil {
		return urls, err
	}

	defer stmt.Close()

	for index, url := range urls {
		var dbURL URL

		err := stmt.QueryRowContext(context, url.ID, url.Short, url.Original, url.UserID).Scan(&dbURL.ID, &dbURL.Short, &dbURL.Original, &dbURL.UserID)
		if err != nil {
			return urls, err
		}

		urls[index] = dbURL
	}

	return urls, tx.Commit()
}

func (r PostgresRepository) Get(context context.Context, id string) (URL, bool) {
	var url URL

	err := r.database.QueryRowContext(context, `SELECT id, short, original, user_id FROM url WHERE id=$1`, id).Scan(&url.ID, &url.Short, &url.Original, &url.UserID)
	if err != nil {
		fmt.Println(err)
		return url, false
	}

	return url, true
}

func (r PostgresRepository) GetAllByUser(context context.Context, userID string) ([]URL, error) {
	var count int
	e := r.database.QueryRowContext(context, `SELECT COUNT(*) FROM url WHERE user_id=$1`, userID).Scan(&count)
	if e != nil {
		return nil, e
	}

	rows, err := r.database.QueryContext(context, `SELECT id, short, original, user_id FROM url WHERE user_id=$1`, userID)
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
