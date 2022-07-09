package entities

import (
	"database/sql"
	"time"
)

type OrderEntity struct {
	Number     string
	Status     string
	Accrual    sql.NullFloat64
	UploadedAt time.Time
}

type UserBalance struct {
	Current          float32
	SummaryWithdrawn float32
}

type UserWithdrawal struct {
	Order       string
	Sum         float32
	ProcessedAt time.Time
}
