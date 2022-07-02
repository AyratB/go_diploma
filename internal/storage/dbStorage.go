package storage

import (
	"context"
	"database/sql"
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
-- 		CREATE TABLE IF NOT EXISTS users (
--     		id 				SERIAL PRIMARY KEY,
--     		guid        	text NOT NULL
-- 		);
-- 		
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
