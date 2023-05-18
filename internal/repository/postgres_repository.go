package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/jackc/pgerrcode"
	"log"
	"os"
	"strings"
	"time"
)

// PostgresRepository is Repository implementation for working with postgesql database.
type PostgresRepository struct {
	database *sql.DB
}

// MakePostgresRepository is constructor for PostgresRepository.
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
	"is_deleted" BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY ("id")
);`)

	if seedError != nil {
		log.Fatal(seedError)
	}

	return repository
}

// Insert adds row in url database table.
func (r PostgresRepository) Insert(ctx context.Context, url URL) (URL, error) {
	_, err := r.database.ExecContext(ctx, `INSERT INTO url (id, short, original, user_id) VALUES ($1, $2, $3, $4);`, url.ID, url.Short, url.Original, url.UserID)

	if err != nil {
		if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
			var dbURL URL
			err := r.database.QueryRowContext(ctx, `SELECT id, short, original, user_id, is_deleted FROM url WHERE id=$1 OR original=$2;`, url.ID, url.Original).Scan(&dbURL.ID, &dbURL.Short, &dbURL.Original, &dbURL.UserID, &dbURL.IsDeleted)
			if err != nil {
				return url, err
			}

			return dbURL, ErrorURLDuplicate
		}
		return url, err
	}

	return url, nil
}

// InsertMany adds many rows in url database table.
func (r PostgresRepository) InsertMany(ctx context.Context, urls []URL) ([]URL, error) {
	tx, err := r.database.Begin()
	if err != nil {
		return urls, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
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

		err := stmt.QueryRowContext(ctx, url.ID, url.Short, url.Original, url.UserID).Scan(&dbURL.ID, &dbURL.Short, &dbURL.Original, &dbURL.UserID)
		if err != nil {
			return urls, err
		}

		urls[index] = dbURL
	}

	return urls, tx.Commit()
}

// Get select row by id from url table.
func (r PostgresRepository) Get(ctx context.Context, id string) (URL, bool) {
	var url URL

	err := r.database.QueryRowContext(ctx, `SELECT id, short, original, user_id, is_deleted FROM url WHERE id=$1`, id).Scan(&url.ID, &url.Short, &url.Original, &url.UserID, &url.IsDeleted)
	if err != nil {
		fmt.Println(err)
		return url, false
	}

	return url, true
}

// GetAllByUser select many rows by user_id from url table.
func (r PostgresRepository) GetAllByUser(ctx context.Context, userID string) ([]URL, error) {
	var count int
	e := r.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM url WHERE user_id=$1 AND is_deleted=false`, userID).Scan(&count)
	if e != nil {
		return nil, e
	}

	rows, err := r.database.QueryContext(ctx, `SELECT id, short, original, user_id, is_deleted FROM url WHERE user_id=$1 AND is_deleted=false`, userID)
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
		err := rows.Scan(&url.ID, &url.Short, &url.Original, &url.UserID, &url.IsDeleted)
		if err != nil {
			return nil, err
		}

		urls[i] = url
		i++
	}

	return urls[:i], nil
}

// DeleteManyByUser delete many rows by user_id in url table.
func (r PostgresRepository) DeleteManyByUser(ctx context.Context, urlIDs []string, userID string) bool {
	param := "{" + strings.Join(urlIDs, ",") + "}"
	_, err := r.database.ExecContext(ctx, `UPDATE url SET is_deleted=true WHERE user_id=$1 AND id=ANY($2::VARCHAR[])`, userID, param)

	return err == nil
}
