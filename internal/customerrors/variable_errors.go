package customerrors

type ErrOrderNumberAlreadyBusy struct {
	OrderUserLogin string
}

func (err ErrOrderNumberAlreadyBusy) Error() string {
	return "Order number is busy by another user"
}
