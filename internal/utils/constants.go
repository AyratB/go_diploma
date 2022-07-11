package utils

type key int
type OrderStatus string

const (
	KeyPrincipalID key = iota
	CookieUserName     = "UserID"

	New        OrderStatus = "NEW"
	Processing OrderStatus = "PROCESSING"
	Invalid    OrderStatus = "INVALID"
	Processed  OrderStatus = "PROCESSED"
)
