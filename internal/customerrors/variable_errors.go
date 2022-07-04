package customerrors

import "fmt"

type ErrOrderNumberAlreadyBusy struct {
	OrderUserLogin string
}

func (err ErrOrderNumberAlreadyBusy) Error() string {
	return "order number is busy by another user"
}

type ErrNoEnoughMoney struct {
	CurrentSum float32
}

func (err ErrNoEnoughMoney) Error() string {
	return fmt.Sprintf("no enough money. Current balance: %v", err.CurrentSum)
}
