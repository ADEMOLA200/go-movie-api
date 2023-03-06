package data

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/showbaba/movies-api/utils"
)

type Comment struct {
	ID           int       `json:"id"`
	MovieID      string    `json:"movie_id"`
	Body         string    `json:"body"`
	UserPublicIP string    `json:"user_public_ip"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

/*
fetch comment by movieID
*/
func (c *Comment) Fetch(movieID string) ([]*Comment, error) {
	rows, err := db.Query(`SELECT id, movie_id, body, user_public_ip, created_at, updated_at FROM comments WHERE movie_id = $1`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*Comment

	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.MovieID, &comment.Body, &comment.UserPublicIP, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

/*
create a new comment
*/
func (c *Comment) Insert() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	var id int
	query := `INSERT INTO comments (movie_id, body, user_public_ip, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err := db.QueryRowContext(ctx, query,
		&c.MovieID, &c.Body,
		&c.UserPublicIP,
		time.Now(), time.Now()).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

/*
this is a custom function that runs db migrations
(on a second thought, i think with some modifications this can be made into a mini package for db migration)
*/
func Migrate() {
	const TOTAL_WORKERS = 1
	var (
		wg      sync.WaitGroup
		errorCh = make(chan error, TOTAL_WORKERS)
	)
	wg.Add(TOTAL_WORKERS)
	log.Println("running db migration :::::::::::::")

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
		defer cancel()
		// check if table exist before creating
		tableExist, err := utils.CheckTableExist(ctx, db, "comments")
		if err != nil {
			errorCh <- err
		}
		if !tableExist {
			query := `CREATE TABLE comments (
			id SERIAL PRIMARY KEY,
			movie_id VARCHAR(255) NOT NULL,
			body VARCHAR(500) NOT NULL,
			user_public_ip VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`
			_, err := db.ExecContext(ctx, query)
			if err != nil {
				errorCh <- err
			}
		}
	}()

	// more go routines can be added here and number of TOTAL_WORKERS increased to handle other tables

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	for err := range errorCh {
		if err != nil {
			panic(err)
		}
	}

	log.Println("complete db migration")
}
