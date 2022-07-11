package repositories

import (
	"github.com/AyratB/go_diploma/internal/entities"
)

type Repository interface {
	RegisterUser(userLogin, password string) error
	LoginUser(userLogin, password string) error
	CheckOrderExists(orderNumber string) error
	SaveOrder(orderNumber, userLogin string) error
	GetUserOrders(userLogin string) ([]entities.OrderEntity, error)
	GetUserBalance(userLogin string) (*entities.UserBalance, error)
	DecreaseBalance(userLogin, order string, sum float32) error
	GetUserWithdrawals(userLogin string) ([]entities.UserWithdrawal, error)
}
