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
