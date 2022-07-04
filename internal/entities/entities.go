package entities

import (
	"database/sql"
	"time"
)

type OrderEntity struct {
	Number     string
	Status     string
	Accrual    sql.NullInt64
	UploadedAt time.Time
}

type UserBalance struct {
	Current   float32
	Withdrawn float32
}
