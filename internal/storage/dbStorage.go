package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AyratB/go_diploma/internal/customerrors"
	"github.com/AyratB/go_diploma/internal/entities"
	"github.com/AyratB/go_diploma/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"time"
)

type DBStorage struct {
	DB *sql.DB
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
     		id 					SERIAL 	PRIMARY KEY,
     		login        		TEXT 	NOT NULL UNIQUE,
     		password 			TEXT 	NOT NULL                                
 		);
 		
 		CREATE TABLE IF NOT EXISTS orders (
     		id 					SERIAL PRIMARY KEY,
    		order_number		TEXT   NOT NULL UNIQUE,
			user_id				INTEGER NOT NULL,
			status				TEXT   NOT NULL,
			accrual				NUMERIC(9,2),
			uploaded_at			TIMESTAMPTZ NOT NULL,
 			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
 		);

		CREATE TABLE IF NOT EXISTS withdrawals (
     		id 						SERIAL PRIMARY KEY,
    		order_number			TEXT   NOT NULL,
			processed_at			TIMESTAMPTZ NOT NULL,
			user_id					INTEGER NOT NULL,
			with_drawn_operation 	NUMERIC(9,2),
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
 		); 
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

func (d *DBStorage) CheckOrderExists(orderNumber string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingOrderUserLogin string

	err := d.DB.QueryRowContext(ctx,
		`SELECT u.login
FROM orders as o
JOIN users as u ON o.user_id = u.id
WHERE o.order_number = $1`, orderNumber).Scan(&existingOrderUserLogin)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if len(existingOrderUserLogin) != 0 {
		return customerrors.ErrOrderNumberAlreadyBusy{
			OrderUserLogin: existingOrderUserLogin,
		}
	}
	return nil
}

func (d *DBStorage) SaveOrder(orderNumber, userLogin string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var userID int
	err := d.DB.
		QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", userLogin).
		Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrInvalidCookie
		}
		return err
	}

	result, err := d.DB.ExecContext(ctx,
		"INSERT INTO orders (order_number, user_id, status, accrual, uploaded_at) VALUES ($1, $2, $3, null, $4)",
		orderNumber, userID, string(utils.New), time.Now())
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

func (d *DBStorage) GetUserOrders(userLogin string) ([]entities.OrderEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := d.DB.QueryContext(ctx,
		`
		SELECT o.order_number, o.status, o.accrual, o.uploaded_at 
			FROM orders as o
		JOIN users as u ON o.user_id = u.id
		WHERE u.login = $1
	`, userLogin)

	defer rows.Close()

	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	orders := make([]entities.OrderEntity, 0)

	for rows.Next() {
		var order entities.OrderEntity
		if err = rows.
			Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (d *DBStorage) GetUserBalance(userLogin string) (userBalance *entities.UserBalance, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var userwithdrawls float32

	err = d.DB.
		QueryRowContext(ctx, `
SELECT SUM(COALESCE( w.with_drawn_operation, 0 ))
FROM users as u
LEFT JOIN withdrawals as w ON w.user_id = u.id
WHERE u.login = $1
GROUP BY u.id`, userLogin).
		Scan(&userwithdrawls)
	if err != nil {
		return
	}

	var userAccruals float32
	err = d.DB.
		QueryRowContext(ctx, `
SELECT SUM(COALESCE( o.accrual, 0 ))
FROM users as u
LEFT JOIN orders as o ON o.user_id = u.id
WHERE u.login = $1
GROUP BY u.id`, userLogin).
		Scan(&userAccruals)
	if err != nil {
		return
	}

	userBalance = &entities.UserBalance{
		Current:          userAccruals - userwithdrawls,
		SummaryWithdrawn: userwithdrawls,
	}

	return
}

func (d *DBStorage) DecreaseBalance(userLogin, orderNumber string, sum float32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var userID int

	if err := d.DB.
		QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", userLogin).
		Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customerrors.ErrInvalidCookie
		}
		return err
	}

	insertResult, err := d.DB.ExecContext(ctx,
		`INSERT INTO withdrawals (order_number, processed_at, user_id, with_drawn_operation) 
VALUES ($1, $2, $3, $4)`, orderNumber, time.Now(), userID, sum)
	if err != nil {
		return err
	}
	rows, err := insertResult.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return nil
}

func (d *DBStorage) GetUserWithdrawals(userLogin string) ([]entities.UserWithdrawal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := d.DB.QueryContext(ctx,
		`
		SELECT w.order_number, w.processed_at, w.with_drawn_operation
			FROM withdrawals as w
		JOIN users as u ON w.user_id = u.id
		WHERE u.login = $1
	`, userLogin)

	defer rows.Close()

	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	userWithdrawals := make([]entities.UserWithdrawal, 0)

	for rows.Next() {
		var userWithdrawal entities.UserWithdrawal
		if err = rows.
			Scan(&userWithdrawal.Order, &userWithdrawal.ProcessedAt, &userWithdrawal.Sum); err != nil {
			return nil, err
		}
		userWithdrawals = append(userWithdrawals, userWithdrawal)
	}
	return userWithdrawals, nil
}

func (d *DBStorage) UpdateOrder(number, status string, accrual *float64) error {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	updateResult, err := d.DB.ExecContext(ctx,
		`UPDATE orders
	SET status = $1,
	 	accrual = $2
	WHERE order_number = $3`, status, accrual, number)
	if err != nil {
		return err
	}
	rows, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return nil
}
