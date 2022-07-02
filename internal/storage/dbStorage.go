package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/AyratB/go_diploma/internal/customerrors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"time"
)

type DBStorage struct {
	DB *sql.DB

	//shortUserURLs map[string]map[string]*entities.URLInfo // key - origin, Value - URLInfo
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	dbStorage := &DBStorage{DB: db}

	if err = dbStorage.initTables(); err != nil {
		return nil, err
	}

	return dbStorage, nil
}

func (d *DBStorage) CloseResources() error {
	//if upStmt != nil {
	//	upStmt.Close()
	//}
	return d.DB.Close()
}

func (d *DBStorage) initTables() error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	initQuery := `
 		CREATE TABLE IF NOT EXISTS users (
     		id 				SERIAL 	PRIMARY KEY,
     		login        	text 	NOT NULL UNIQUE,
     		password 		text 	NOT NULL
 		);
 		
-- 		CREATE TABLE IF NOT EXISTS user_urls (
--     		id 				SERIAL PRIMARY KEY,
--     		original_url 	TEXT NOT NULL,
-- 			shorten_url		TEXT NOT NULL,
-- 			user_id			INTEGER NOT NULL,
-- 			is_deleted		BOOLEAN DEFAULT false,
-- 			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
-- 		);
-- 
-- 		CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON user_urls (original_url);
	`
	if _, err := d.DB.ExecContext(ctx, initQuery); err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) RegisterUser(login, password string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := d.DB.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)", login, password)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == pgerrcode.UniqueViolation {
			return customerrors.ErrDuplicateUserLogin
		}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return nil
}

func (d *DBStorage) LoginUser(login, password string) error {

	var isUserExist bool

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := d.DB.
		QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 from users WHERE login = $1 AND password = $2)", login, password).
		Scan(&isUserExist)
	if err != nil {
		return err
	}
	if !isUserExist {
		return customerrors.ErrNoUserByLoginAndPassword
	}
	return nil
}
